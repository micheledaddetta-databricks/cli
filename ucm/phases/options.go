package phases

import (
	"context"

	"github.com/databricks/cli/ucm"
	"github.com/databricks/cli/ucm/deploy"
	"github.com/databricks/cli/ucm/deploy/terraform"
)

// TerraformWrapper is the subset of *terraform.Terraform that phases depend on.
// Keeping the surface minimal lets tests inject a fake without standing up a
// real terraform binary. *terraform.Terraform satisfies this interface so the
// production factory does not need an adapter.
type TerraformWrapper interface {
	Render(ctx context.Context, u *ucm.Ucm) error
	Init(ctx context.Context, u *ucm.Ucm) error
	Plan(ctx context.Context, u *ucm.Ucm) (*terraform.PlanResult, error)
	Apply(ctx context.Context, u *ucm.Ucm, forceLock bool) error
	Destroy(ctx context.Context, u *ucm.Ucm, forceLock bool) error
	Import(ctx context.Context, u *ucm.Ucm, address, id string) error
	StateRm(ctx context.Context, u *ucm.Ucm, address string) error
}

// TerraformFactory constructs a terraform-engine wrapper scoped to u.
// Production callers pass DefaultTerraformFactory; tests hand in a factory
// that returns a fake.
type TerraformFactory func(ctx context.Context, u *ucm.Ucm) (TerraformWrapper, error)

// Compile-time assertion that *terraform.Terraform satisfies TerraformWrapper.
// Keeps the interface honest when the underlying wrapper gains new methods;
// a broken assertion catches the drift at build time rather than at the
// DefaultTerraformFactory call site.
var _ TerraformWrapper = (*terraform.Terraform)(nil)

// DefaultTerraformFactory builds a real *terraform.Terraform via terraform.New,
// resolving (and if necessary downloading) the terraform binary on first use.
// Production callers pass this directly; tests never should.
func DefaultTerraformFactory(ctx context.Context, u *ucm.Ucm) (TerraformWrapper, error) {
	return terraform.New(ctx, u)
}

// Options bundles the externally-supplied dependencies a phase needs at
// runtime. The CLI layer populates Backend and the engine factory before
// invoking plan/deploy/destroy; tests may omit Backend when exercising the
// direct engine or the no-op initialize error paths.
type Options struct {
	// Backend is the pull/push state-storage pair used by Initialize and
	// the post-apply/destroy Push. Required for Plan/Deploy/Destroy in the
	// terraform engine; direct-engine callers may leave it nil since there
	// is no remote state to pull.
	Backend deploy.Backend

	// TerraformFactory produces the terraform wrapper bound to u. When nil,
	// phases fall back to DefaultTerraformFactory.
	TerraformFactory TerraformFactory

	// ForceLock mirrors the --force-lock flag: when true, Pull/Push and
	// terraform Apply/Destroy override an existing deploy lock instead of
	// failing with ErrLockHeld.
	ForceLock bool

	// AutoApprove mirrors the --auto-approve flag: when true, Deploy and
	// Destroy skip the destructive-action prompt. Non-destructive plans
	// never prompt regardless of this value.
	AutoApprove bool
}

// terraformFactoryOrDefault returns o.TerraformFactory or the production
// factory when unset.
func (o Options) terraformFactoryOrDefault() TerraformFactory {
	if o.TerraformFactory != nil {
		return o.TerraformFactory
	}
	return DefaultTerraformFactory
}
