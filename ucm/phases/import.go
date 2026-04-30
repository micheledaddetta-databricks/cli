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

// ImportKind identifies the resource kind an Import call targets.
// The string value matches the CLI argument the user types.
type ImportKind string

const (
	ImportCatalog           ImportKind = "catalog"
	ImportSchema            ImportKind = "schema"
	ImportStorageCredential ImportKind = "storage_credential"
	ImportExternalLocation  ImportKind = "external_location"
	ImportVolume            ImportKind = "volume"
	ImportConnection        ImportKind = "connection"
)

// ImportRequest bundles the operator-supplied inputs for a single import.
// Name is the UC identifier (e.g. "sales_prod" for a catalog, "sales.raw"
// for a schema); Key is the ucm.yml map key the imported object will be
// recorded under.
type ImportRequest struct {
	Kind ImportKind
	Name string
	Key  string
}

// Import resolves the deployment engine and dispatches to the direct or
// terraform implementation. Errors are reported via logdiag; callers must
// check logdiag.HasError before continuing. The terraform path pushes state
// on success; the direct path rewrites resources.json in place.
func Import(ctx context.Context, u *ucm.Ucm, opts Options, req ImportRequest) {
	log.Infof(ctx, "Phase: import %s %s", req.Kind, req.Name)

	setting := Initialize(ctx, u, opts)
	if logdiag.HasError(ctx) {
		return
	}

	if err := validateResourceDeclared(u, req); err != nil {
		logdiag.LogError(ctx, err)
		return
	}

	if setting.Type.IsDirect() {
		importDirect(ctx, u, opts, req)
		return
	}
	importTerraform(ctx, u, opts, req)
}

// validateResourceDeclared errors out when the ucm.yml map for the given
// kind does not contain the requested key. Import is a bind-to-existing
// operation — it refuses to seed state for an undeclared resource.
func validateResourceDeclared(u *ucm.Ucm, req ImportRequest) error {
	declared := false
	switch req.Kind {
	case ImportCatalog:
		_, declared = u.Config.Resources.Catalogs[req.Key]
	case ImportSchema:
		_, declared = u.Config.Resources.Schemas[req.Key]
	case ImportStorageCredential:
		_, declared = u.Config.Resources.StorageCredentials[req.Key]
	case ImportExternalLocation:
		_, declared = u.Config.Resources.ExternalLocations[req.Key]
	case ImportVolume:
		_, declared = u.Config.Resources.Volumes[req.Key]
	case ImportConnection:
		_, declared = u.Config.Resources.Connections[req.Key]
	default:
		return fmt.Errorf("ucm import: unsupported kind %q", req.Kind)
	}
	if !declared {
		return fmt.Errorf("ucm import: resources.%s.%s is not declared in ucm.yml — "+
			"run `ucm import` only after declaring the resource in ucm.yml; "+
			"then ucm import will bind state to the existing UC object", pluralKind(req.Kind), req.Key)
	}
	return nil
}

// pluralKind maps ImportKind to the ucm.yml map name under resources.
// Kept local so the CLI layer never has to spell these strings out.
func pluralKind(k ImportKind) string {
	switch k {
	case ImportCatalog:
		return "catalogs"
	case ImportSchema:
		return "schemas"
	case ImportStorageCredential:
		return "storage_credentials"
	case ImportExternalLocation:
		return "external_locations"
	case ImportVolume:
		return "volumes"
	case ImportConnection:
		return "connections"
	}
	return string(k)
}

// terraformAddress builds the `databricks_<type>.<key>` address the terraform
// provider expects for the given resource.
func terraformAddress(req ImportRequest) string {
	return "databricks_" + string(req.Kind) + "." + req.Key
}

func importTerraform(ctx context.Context, u *ucm.Ucm, opts Options, req ImportRequest) {
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

	if err := tf.Import(ctx, u, terraformAddress(req), req.Name); err != nil {
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

// importDirect resolves the dresources adapter for the requested kind, reads
// the live UC object via the SDK, RemapState's it into the saved-state shape
// that ucm/direct.Apply persists on a normal create, and writes it through
// dstate.DeploymentState.SaveState. Refuses to overwrite an entry that is
// already bound — operators must `ucm deployment unbind` first to make the
// imported state match a re-discovered live object.
func importDirect(ctx context.Context, u *ucm.Ucm, _ Options, req ImportRequest) {
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
		logdiag.LogError(ctx, fmt.Errorf("ucm import: no adapter for kind %q", req.Kind))
		return
	}

	var db dstate.DeploymentState
	if err := db.Open(DirectStatePath(u)); err != nil {
		logdiag.LogError(ctx, fmt.Errorf("open direct state: %w", err))
		return
	}

	stateKey := fmt.Sprintf("resources.%s.%s", plural, req.Key)
	if _, exists := db.Data.State[stateKey]; exists {
		logdiag.LogError(ctx, fmt.Errorf("ucm import: %s is already bound in state — use `ucm deployment unbind` first", stateKey))
		return
	}

	live, err := adapter.DoRead(ctx, req.Name)
	if err != nil {
		if errors.Is(err, apierr.ErrResourceDoesNotExist) || errors.Is(err, apierr.ErrNotFound) {
			logdiag.LogError(ctx, fmt.Errorf("ucm import: %s %q not found in Unity Catalog", req.Kind, req.Name))
			return
		}
		logdiag.LogError(ctx, fmt.Errorf("read %s %q: %w", req.Kind, req.Name, err))
		return
	}
	if live == nil {
		logdiag.LogError(ctx, fmt.Errorf("ucm import: %s %q not found in Unity Catalog", req.Kind, req.Name))
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

	log.Infof(ctx, "direct: imported %s %s as %s", req.Kind, req.Name, stateKey)
}
