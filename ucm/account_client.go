package ucm

import (
	"fmt"
	"sync"

	"github.com/databricks/cli/libs/databrickscfg"
	"github.com/databricks/databricks-sdk-go"
	sdkconfig "github.com/databricks/databricks-sdk-go/config"
)

// accountClientConfig builds the SDK config used to construct an account
// client for this Ucm. The account host is sourced from
// Config.Workspace.AccountHost (workspace Host alone is insufficient because
// the SDK routes account vs workspace by host).
func (u *Ucm) accountClientConfig() *sdkconfig.Config {
	return &sdkconfig.Config{
		Host:    u.Config.Workspace.AccountHost,
		Profile: u.Config.Workspace.Profile,
	}
}

// buildAccountClient resolves auth configuration and constructs an account
// client. Mirrors buildWorkspaceClient: when only the host is set, install
// the ResolveProfileFromHost loader so the SDK picks up a unique matching
// profile from ~/.databrickscfg. On ambiguity the loader returns the
// errMultipleProfiles error detected by databrickscfg.AsMultipleProfiles.
func (u *Ucm) buildAccountClient() (*databricks.AccountClient, error) {
	cfg := u.accountClientConfig()

	if cfg.Host != "" && cfg.Profile == "" {
		cfg.Loaders = []sdkconfig.Loader{
			sdkconfig.ConfigAttributes,
			databrickscfg.ResolveProfileFromHost,
		}
	}

	if err := cfg.EnsureResolved(); err != nil {
		return nil, err
	}

	if cfg.Host != "" && cfg.Profile != "" {
		if err := databrickscfg.ValidateConfigAndProfileHost(cfg, cfg.Profile); err != nil {
			return nil, err
		}
	}

	return databricks.NewAccountClient((*databricks.Config)(cfg))
}

func (u *Ucm) initAccountClientOnce() {
	u.getAccountClient = sync.OnceValues(func() (*databricks.AccountClient, error) {
		a, err := u.buildAccountClient()
		if err != nil {
			return nil, fmt.Errorf("cannot resolve ucm account auth configuration: %w", err)
		}
		return a, nil
	})
}

// AccountClientE returns the memoized account client, building it from
// Config.Workspace on first call.
func (u *Ucm) AccountClientE() (*databricks.AccountClient, error) {
	if u.getAccountClient == nil {
		u.initAccountClientOnce()
	}
	return u.getAccountClient()
}

// AccountClient is the panicking convenience wrapper around AccountClientE.
// Prefer AccountClientE in new code so callers can surface auth errors.
func (u *Ucm) AccountClient() *databricks.AccountClient {
	client, err := u.AccountClientE()
	if err != nil {
		panic(err)
	}
	return client
}

// SetAccountClient injects a pre-built client, primarily for tests.
func (u *Ucm) SetAccountClient(a *databricks.AccountClient) {
	u.getAccountClient = func() (*databricks.AccountClient, error) {
		return a, nil
	}
}

// ClearAccountClient resets the memoized client so the next AccountClientE
// call rebuilds it. Used after Config.Workspace is mutated (e.g. when a
// profile is selected via the ambiguity picker).
func (u *Ucm) ClearAccountClient() {
	u.initAccountClientOnce()
}
