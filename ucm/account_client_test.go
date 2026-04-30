package ucm

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/databricks/cli/internal/testutil"
	"github.com/databricks/cli/ucm/config"
	"github.com/databricks/databricks-sdk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupAccountDatabricksCfg writes a .databrickscfg with account-shaped
// profiles and points the SDK at it. Account profiles are matched by the
// presence of account_id in the SDK profile loader.
func setupAccountDatabricksCfg(t *testing.T) {
	t.Helper()
	tempHomeDir := t.TempDir()
	homeEnvVar := "HOME"
	if runtime.GOOS == "windows" {
		homeEnvVar = "USERPROFILE"
	}

	cfg := []byte(strings.Join([]string{
		"[ACCT-UNIQUE]",
		"host = https://accounts-unique.cloud.databricks.com",
		"account_id = 11111111-1111-1111-1111-111111111111",
		"token = a",
		"",
		"[ACCT-OTHER]",
		"host = https://accounts-other.cloud.databricks.com",
		"account_id = 22222222-2222-2222-2222-222222222222",
		"token = b",
		"",
	}, "\n"))
	err := os.WriteFile(filepath.Join(tempHomeDir, ".databrickscfg"), cfg, 0o644)
	require.NoError(t, err)

	t.Setenv("DATABRICKS_CONFIG_FILE", "")
	t.Setenv(homeEnvVar, tempHomeDir)
}

func TestAccountClientE_BuildsFromConfig(t *testing.T) {
	testutil.CleanupEnvironment(t)
	setupAccountDatabricksCfg(t)

	u := &Ucm{
		Config: config.Root{
			Workspace: config.Workspace{
				AccountHost: "https://accounts-unique.cloud.databricks.com",
			},
		},
	}

	client, err := u.AccountClientE()
	require.NoError(t, err)
	assert.Equal(t, "ACCT-UNIQUE", client.Config.Profile)
	assert.Equal(t, "https://accounts-unique.cloud.databricks.com", client.Config.Host)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", client.Config.AccountID)
}

func TestAccountClientE_Memoizes(t *testing.T) {
	testutil.CleanupEnvironment(t)
	setupAccountDatabricksCfg(t)

	u := &Ucm{
		Config: config.Root{
			Workspace: config.Workspace{
				AccountHost: "https://accounts-unique.cloud.databricks.com",
			},
		},
	}

	c1, err := u.AccountClientE()
	require.NoError(t, err)
	c2, err := u.AccountClientE()
	require.NoError(t, err)
	assert.Same(t, c1, c2, "expected AccountClientE to memoize")
}

func TestAccountClientE_ClearReResolves(t *testing.T) {
	testutil.CleanupEnvironment(t)
	setupAccountDatabricksCfg(t)

	u := &Ucm{
		Config: config.Root{
			Workspace: config.Workspace{
				AccountHost: "https://accounts-unique.cloud.databricks.com",
			},
		},
	}

	c1, err := u.AccountClientE()
	require.NoError(t, err)

	u.ClearAccountClient()
	c2, err := u.AccountClientE()
	require.NoError(t, err)
	assert.NotSame(t, c1, c2, "expected ClearAccountClient to reset memoization")
}

func TestSetAccountClient_OverridesBuilder(t *testing.T) {
	u := &Ucm{}
	injected := &databricks.AccountClient{}
	u.SetAccountClient(injected)

	got, err := u.AccountClientE()
	require.NoError(t, err)
	assert.Same(t, injected, got, "expected SetAccountClient to inject the given client")
}
