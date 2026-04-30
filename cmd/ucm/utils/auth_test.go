package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/databricks/cli/cmd/root"
	"github.com/databricks/cli/internal/testutil"
	"github.com/databricks/cli/libs/cmdctx"
	"github.com/databricks/cli/libs/cmdio"
	"github.com/databricks/cli/libs/logdiag"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupDatabricksCfg mirrors the helper in ucm/workspace_client_test.go.
func setupDatabricksCfg(t *testing.T) {
	t.Helper()
	tempHomeDir := t.TempDir()
	homeEnvVar := "HOME"
	if runtime.GOOS == "windows" {
		homeEnvVar = "USERPROFILE"
	}

	cfg := []byte(strings.Join([]string{
		"[PROFILE-UNIQUE]",
		"host = https://unique.example.com",
		"token = u",
		"",
		"[PROFILE-DUP-1]",
		"host = https://dup.example.com",
		"token = d1",
		"",
		"[PROFILE-DUP-2]",
		"host = https://dup.example.com",
		"token = d2",
		"",
	}, "\n"))
	err := os.WriteFile(filepath.Join(tempHomeDir, ".databrickscfg"), cfg, 0o644)
	require.NoError(t, err)

	t.Setenv("DATABRICKS_CONFIG_FILE", "")
	t.Setenv(homeEnvVar, tempHomeDir)
}

// setupUcmWithHost drops a ucm.yml with the given host into a fresh working
// directory and returns a *cobra.Command wired for MustWorkspaceClient.
//
// MustWorkspaceClient owns logdiag.InitContext (faithful fork from upstream
// cmd/root/auth.go); pre-initializing here would panic. Tests assert against
// logdiag's HasError/GetFirstErrorSummary on the post-init context the
// function installs back on cmd.
func setupUcmWithHost(t *testing.T, host string) *cobra.Command {
	t.Helper()
	rootPath := t.TempDir()
	t.Chdir(rootPath)

	y := "ucm:\n  name: t\n\nworkspace:\n  host: " + host + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(rootPath, "ucm.yml"), []byte(y), 0o644))

	cmd := &cobra.Command{Use: "plan"}
	cmd.PersistentFlags().String("target", "", "")
	cmd.PersistentFlags().String("profile", "", "")

	cmd.SetContext(cmdio.MockDiscard(t.Context()))
	return cmd
}

func TestMustWorkspaceClient_UniqueHostMatchResolvesProfile(t *testing.T) {
	testutil.CleanupEnvironment(t)
	// Restrict PATH so the SDK's auth resolution cannot invoke az/gcloud.
	if runtime.GOOS == "windows" {
		t.Setenv("PATH", `C:\Windows\System32`)
	} else {
		t.Setenv("PATH", "/usr/bin:/bin")
	}
	setupDatabricksCfg(t)

	cmd := setupUcmWithHost(t, "https://unique.example.com")
	err := MustWorkspaceClient(cmd, nil)

	require.NoError(t, err)
	require.False(t, logdiag.HasError(cmd.Context()))
	w := cmdctx.WorkspaceClient(cmd.Context())
	assert.Equal(t, "PROFILE-UNIQUE", w.Config.Profile)
	assert.Equal(t, "https://unique.example.com", w.Config.Host)
}

func TestMustWorkspaceClient_AmbiguousHostReturnsGuidanceError(t *testing.T) {
	testutil.CleanupEnvironment(t)
	if runtime.GOOS == "windows" {
		t.Setenv("PATH", `C:\Windows\System32`)
	} else {
		t.Setenv("PATH", "/usr/bin:/bin")
	}
	setupDatabricksCfg(t)

	cmd := setupUcmWithHost(t, "https://dup.example.com")
	err := MustWorkspaceClient(cmd, nil)

	require.ErrorIs(t, err, root.ErrAlreadyPrinted)
	require.True(t, logdiag.HasError(cmd.Context()))
	summary := logdiag.GetFirstErrorSummary(cmd.Context())
	assert.Contains(t, summary, "multiple profiles matched")
	assert.Contains(t, summary, "Matching workspace profiles")
	assert.Contains(t, summary, "PROFILE-DUP-1")
	assert.Contains(t, summary, "PROFILE-DUP-2")
}

func TestMustWorkspaceClient_NoMatchingProfileDoesNotPrompt(t *testing.T) {
	testutil.CleanupEnvironment(t)
	if runtime.GOOS == "windows" {
		t.Setenv("PATH", `C:\Windows\System32`)
	} else {
		t.Setenv("PATH", "/usr/bin:/bin")
	}
	setupDatabricksCfg(t)

	cmd := setupUcmWithHost(t, "https://nobody.example.com")
	err := MustWorkspaceClient(cmd, nil)

	// DAB-parallel behavior: the ResolveProfileFromHost loader swallows
	// "no matching profiles" so EnsureResolved may succeed with only the host.
	// The critical invariant is that we never surface the "multi-match"
	// ambiguity text and never drop into the interactive picker. Either a
	// clean error or a clean success is acceptable.
	if err != nil {
		assert.NotContains(t, err.Error(), "Multiple profiles")
		assert.NotContains(t, err.Error(), "multiple profiles matched")
		summary := logdiag.GetFirstErrorSummary(cmd.Context())
		assert.NotContains(t, summary, "Multiple profiles")
		assert.NotContains(t, summary, "multiple profiles matched")
		return
	}
	assert.False(t, logdiag.HasError(cmd.Context()))
}

func TestMustWorkspaceClient_ProfileInYamlUsedVerbatim(t *testing.T) {
	testutil.CleanupEnvironment(t)
	if runtime.GOOS == "windows" {
		t.Setenv("PATH", `C:\Windows\System32`)
	} else {
		t.Setenv("PATH", "/usr/bin:/bin")
	}
	setupDatabricksCfg(t)

	rootPath := t.TempDir()
	t.Chdir(rootPath)

	y := "ucm:\n  name: t\n\nworkspace:\n  host: https://unique.example.com\n  profile: PROFILE-UNIQUE\n"
	require.NoError(t, os.WriteFile(filepath.Join(rootPath, "ucm.yml"), []byte(y), 0o644))

	cmd := &cobra.Command{Use: "plan"}
	cmd.PersistentFlags().String("target", "", "")
	cmd.PersistentFlags().String("profile", "", "")
	cmd.SetContext(cmdio.MockDiscard(t.Context()))

	err := MustWorkspaceClient(cmd, nil)

	require.NoError(t, err)
	require.False(t, logdiag.HasError(cmd.Context()))
	w := cmdctx.WorkspaceClient(cmd.Context())
	assert.Equal(t, "PROFILE-UNIQUE", w.Config.Profile)
}

// setupAccountDatabricksCfg writes a .databrickscfg with one workspace profile
// (used by TryConfigureUcm's workspace pass) and account profiles for the
// account-side resolution.
func setupAccountDatabricksCfg(t *testing.T) {
	t.Helper()
	tempHomeDir := t.TempDir()
	homeEnvVar := "HOME"
	if runtime.GOOS == "windows" {
		homeEnvVar = "USERPROFILE"
	}

	cfg := []byte(strings.Join([]string{
		"[WS-UNIQUE]",
		"host = https://ws-unique.example.com",
		"token = w",
		"",
		"[ACCT-UNIQUE]",
		"host = https://accounts-unique.cloud.databricks.com",
		"account_id = 11111111-1111-1111-1111-111111111111",
		"token = a",
		"",
	}, "\n"))
	err := os.WriteFile(filepath.Join(tempHomeDir, ".databrickscfg"), cfg, 0o644)
	require.NoError(t, err)

	t.Setenv("DATABRICKS_CONFIG_FILE", "")
	t.Setenv(homeEnvVar, tempHomeDir)
}

func TestMustAccountClient_ResolvesFromYamlAccountHost(t *testing.T) {
	testutil.CleanupEnvironment(t)
	if runtime.GOOS == "windows" {
		t.Setenv("PATH", `C:\Windows\System32`)
	} else {
		t.Setenv("PATH", "/usr/bin:/bin")
	}
	setupAccountDatabricksCfg(t)

	rootPath := t.TempDir()
	t.Chdir(rootPath)

	y := "ucm:\n  name: t\n\nworkspace:\n" +
		"  host: https://ws-unique.example.com\n" +
		"  account_host: https://accounts-unique.cloud.databricks.com\n"
	require.NoError(t, os.WriteFile(filepath.Join(rootPath, "ucm.yml"), []byte(y), 0o644))

	cmd := &cobra.Command{Use: "metastore"}
	cmd.PersistentFlags().String("target", "", "")
	cmd.PersistentFlags().String("profile", "", "")
	cmd.SetContext(cmdio.MockDiscard(t.Context()))

	err := MustAccountClient(cmd, nil)

	require.NoError(t, err)
	require.False(t, logdiag.HasError(cmd.Context()))
	a := cmdctx.AccountClient(cmd.Context())
	assert.Equal(t, "ACCT-UNIQUE", a.Config.Profile)
	assert.Equal(t, "https://accounts-unique.cloud.databricks.com", a.Config.Host)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", a.Config.AccountID)
}

func TestMustAccountClient_NoAccountProfileSurfacesError(t *testing.T) {
	testutil.CleanupEnvironment(t)
	if runtime.GOOS == "windows" {
		t.Setenv("PATH", `C:\Windows\System32`)
	} else {
		t.Setenv("PATH", "/usr/bin:/bin")
	}
	// Use the workspace-only databrickscfg helper: there are no account
	// profiles to resolve, so MustAccountClient must surface a clean error
	// rather than dropping into the picker (PATH is locked down so external
	// auth helpers can't rescue the resolution either).
	setupDatabricksCfg(t)

	rootPath := t.TempDir()
	t.Chdir(rootPath)

	// No account_host configured and no account profiles in databrickscfg.
	y := "ucm:\n  name: t\n\nworkspace:\n  host: https://nobody.example.com\n"
	require.NoError(t, os.WriteFile(filepath.Join(rootPath, "ucm.yml"), []byte(y), 0o644))

	cmd := &cobra.Command{Use: "metastore"}
	cmd.PersistentFlags().String("target", "", "")
	cmd.PersistentFlags().String("profile", "", "")
	cmd.SetContext(cmdio.MockDiscard(t.Context()))

	err := MustAccountClient(cmd, nil)
	require.Error(t, err)
}
