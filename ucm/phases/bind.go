package phases

import (
	"context"
	"errors"
	"fmt"

	"github.com/databricks/cli/libs/log"
	"github.com/databricks/cli/libs/logdiag"
	"github.com/databricks/cli/ucm"
	"github.com/databricks/cli/ucm/config/mutator"
	"github.com/databricks/cli/ucm/deploy"
	"github.com/databricks/cli/ucm/direct/dresources"
	"github.com/databricks/cli/ucm/direct/dstate"
	"github.com/databricks/databricks-sdk-go/apierr"
)

// BindRequest bundles the operator-supplied inputs for a single bind. Name is
// the UC identifier of the existing object (e.g. "team_alpha" for a catalog,
// "team_alpha.bronze" for a schema); Key is the ucm.yml map key the binding
// will be recorded under. Kind reuses the ImportKind vocabulary since bind
// and import share the same per-resource primitives.
type BindRequest struct {
	Kind ImportKind
	Name string
	Key  string
}

// UnbindRequest mirrors BindRequest but only needs the Kind+Key pair — unbind
// drops the recorded state entry without touching the remote UC object, so
// the UC name is immaterial.
type UnbindRequest struct {
	Kind ImportKind
	Key  string
}

// Bind resolves the deployment engine and attaches an existing Unity Catalog
// object to the ucm-declared key in req.Key. The direct engine records a
// state entry; the terraform engine runs `terraform import`. Errors are
// reported via logdiag; callers must check logdiag.HasError before
// continuing. The terraform path pushes state on success; the direct path
// rewrites resources.json in place. Mirrors bundle/phases/bind.go in shape.
func Bind(ctx context.Context, u *ucm.Ucm, opts Options, req BindRequest) {
	log.Infof(ctx, "Phase: bind %s %s", req.Kind, req.Name)

	setting := Initialize(ctx, u, opts)
	if logdiag.HasError(ctx) {
		return
	}

	if err := validateResourceDeclared(u, ImportRequest(req)); err != nil {
		logdiag.LogError(ctx, err)
		return
	}

	if setting.Type.IsDirect() {
		bindDirect(ctx, u, opts, req)
		return
	}
	bindTerraform(ctx, u, opts, req)
}

// Unbind resolves the deployment engine and drops the recorded binding for
// req.Key. The direct engine deletes the state entry in resources.json; the
// terraform engine runs `terraform state rm`. The remote UC object is never
// touched — unbind is a state-only operation. Mirrors
// bundle/phases/bind.go::Unbind in shape.
func Unbind(ctx context.Context, u *ucm.Ucm, opts Options, req UnbindRequest) {
	log.Infof(ctx, "Phase: unbind %s %s", req.Kind, req.Key)

	setting := Initialize(ctx, u, opts)
	if logdiag.HasError(ctx) {
		return
	}

	if setting.Type.IsDirect() {
		unbindDirect(ctx, u, opts, req)
		return
	}
	unbindTerraform(ctx, u, opts, req)
}

func bindTerraform(ctx context.Context, u *ucm.Ucm, opts Options, req BindRequest) {
	factory := opts.terraformFactoryOrDefault()
	tf, err := factory(ctx, u)
	if err != nil {
		logdiag.LogError(ctx, fmt.Errorf("build terraform wrapper: %w", err))
		return
	}

	if err := tf.Render(ctx, u); err != nil {
		logdiag.LogError(ctx, fmt.Errorf("render terraform config: %w", err))
		return
	}

	if err := tf.Init(ctx, u); err != nil {
		logdiag.LogError(ctx, fmt.Errorf("terraform init: %w", err))
		return
	}

	if err := tf.Import(ctx, u, terraformAddress(ImportRequest(req)), req.Name); err != nil {
		logdiag.LogError(ctx, fmt.Errorf("terraform import: %w", err))
		return
	}

	// StateUpdate must run before Push so the pushed blob carries a fresh
	// Seq/CliVersion/Timestamp/UUID. Push only mirrors local.
	ucm.ApplyContext(ctx, u, deploy.StateUpdate())
	if logdiag.HasError(ctx) {
		return
	}

	if err := deploy.Push(ctx, u, opts.Backend); err != nil {
		logdiag.LogError(ctx, fmt.Errorf("push remote state: %w", err))
		return
	}
}

// bindDirect resolves the dresources adapter for the requested kind, reads
// the live UC object via the SDK, RemapState's it into the saved-state shape
// that ucm/direct.Apply persists on a normal create, and writes it through
// dstate.DeploymentState.SaveState. Refuses to overwrite an entry that is
// already bound — operators must `ucm deployment unbind` first to rebind a
// re-discovered live object. Mirrors importDirect; bind and import share the
// same per-resource primitives.
func bindDirect(ctx context.Context, u *ucm.Ucm, _ Options, req BindRequest) {
	ucm.ApplyContext(ctx, u, mutator.ResolveVariableReferencesOnlyResources("resources"))
	if logdiag.HasError(ctx) {
		return
	}

	client, err := u.WorkspaceClientE()
	if err != nil {
		logdiag.LogError(ctx, fmt.Errorf("resolve workspace client: %w", err))
		return
	}

	adapters, err := dresources.InitAll(client)
	if err != nil {
		logdiag.LogError(ctx, fmt.Errorf("init resource adapters: %w", err))
		return
	}
	plural := pluralKind(req.Kind)
	adapter, ok := adapters[plural]
	if !ok {
		logdiag.LogError(ctx, fmt.Errorf("ucm bind: no adapter for kind %q", req.Kind))
		return
	}

	var db dstate.DeploymentState
	if err := db.Open(DirectStatePath(u)); err != nil {
		logdiag.LogError(ctx, fmt.Errorf("open direct state: %w", err))
		return
	}

	stateKey := fmt.Sprintf("resources.%s.%s", plural, req.Key)
	if _, exists := db.Data.State[stateKey]; exists {
		logdiag.LogError(ctx, fmt.Errorf("ucm bind: %s is already bound in state — use `ucm deployment unbind` first", stateKey))
		return
	}

	live, err := adapter.DoRead(ctx, req.Name)
	if err != nil {
		if errors.Is(err, apierr.ErrResourceDoesNotExist) || errors.Is(err, apierr.ErrNotFound) {
			logdiag.LogError(ctx, fmt.Errorf("ucm bind: %s %q not found in Unity Catalog", req.Kind, req.Name))
			return
		}
		logdiag.LogError(ctx, fmt.Errorf("read %s %q: %w", req.Kind, req.Name, err))
		return
	}
	if live == nil {
		logdiag.LogError(ctx, fmt.Errorf("ucm bind: %s %q not found in Unity Catalog", req.Kind, req.Name))
		return
	}

	saved, err := adapter.RemapState(live)
	if err != nil {
		logdiag.LogError(ctx, fmt.Errorf("remap %s state: %w", req.Kind, err))
		return
	}

	if err := db.SaveState(stateKey, req.Name, saved, nil); err != nil {
		logdiag.LogError(ctx, fmt.Errorf("save state for %s: %w", stateKey, err))
		return
	}

	if err := db.Finalize(); err != nil {
		logdiag.LogError(ctx, fmt.Errorf("finalize direct state: %w", err))
		return
	}

	log.Infof(ctx, "direct: bound %s %s as %s", req.Kind, req.Name, stateKey)
}

func unbindTerraform(ctx context.Context, u *ucm.Ucm, opts Options, req UnbindRequest) {
	factory := opts.terraformFactoryOrDefault()
	tf, err := factory(ctx, u)
	if err != nil {
		logdiag.LogError(ctx, fmt.Errorf("build terraform wrapper: %w", err))
		return
	}

	if err := tf.Render(ctx, u); err != nil {
		logdiag.LogError(ctx, fmt.Errorf("render terraform config: %w", err))
		return
	}

	if err := tf.Init(ctx, u); err != nil {
		logdiag.LogError(ctx, fmt.Errorf("terraform init: %w", err))
		return
	}

	address := terraformAddress(ImportRequest{Kind: req.Kind, Key: req.Key})
	if err := tf.StateRm(ctx, u, address); err != nil {
		logdiag.LogError(ctx, fmt.Errorf("terraform state rm: %w", err))
		return
	}

	// StateUpdate must run before Push so the pushed blob carries a fresh
	// Seq/CliVersion/Timestamp/UUID. Push only mirrors local.
	ucm.ApplyContext(ctx, u, deploy.StateUpdate())
	if logdiag.HasError(ctx) {
		return
	}

	if err := deploy.Push(ctx, u, opts.Backend); err != nil {
		logdiag.LogError(ctx, fmt.Errorf("push remote state: %w", err))
		return
	}
}

// unbindDirect drops the recorded state entry for req.Key without touching
// the remote UC object. It is a state-only operation: open the database,
// guard against a missing key, DeleteState, Finalize. The Initialize step
// has already validated that the engine is direct.
func unbindDirect(ctx context.Context, u *ucm.Ucm, _ Options, req UnbindRequest) {
	var db dstate.DeploymentState
	if err := db.Open(DirectStatePath(u)); err != nil {
		logdiag.LogError(ctx, fmt.Errorf("open direct state: %w", err))
		return
	}

	stateKey := fmt.Sprintf("resources.%s.%s", pluralKind(req.Kind), req.Key)
	if _, exists := db.Data.State[stateKey]; !exists {
		logdiag.LogError(ctx, fmt.Errorf("ucm unbind: %s is not bound in state", stateKey))
		return
	}

	if err := db.DeleteState(stateKey); err != nil {
		logdiag.LogError(ctx, fmt.Errorf("delete state entry %s: %w", stateKey, err))
		return
	}

	if err := db.Finalize(); err != nil {
		logdiag.LogError(ctx, fmt.Errorf("finalize direct state: %w", err))
		return
	}

	log.Infof(ctx, "direct: unbound %s", stateKey)
}
