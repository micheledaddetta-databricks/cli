package deployment

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/databricks/cli/cmd/root"
	"github.com/databricks/cli/libs/cmdio"
	"github.com/databricks/cli/ucm/deploy"
	"github.com/databricks/cli/ucm/phases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMigrate_HelpTextIsUcmFlavored guards against accidental "bundle" leaks
// in the migrate verb's user-facing prose. The full bundle source mentions
// "bundle" in its help text; the ucm fork must reword every occurrence so
// `databricks ucm deployment migrate --help` reads cleanly.
func TestMigrate_HelpTextIsUcmFlavored(t *testing.T) {
	cmd := newMigrateCommand()

	short := cmd.Short
	long := cmd.Long

	// Short summary should describe the engine migration without the
	// "bundle" framing.
	assert.NotEmpty(t, short)
	assert.NotContains(t, strings.ToLower(short), "bundle")

	// Long help should not advertise `bundle deploy` / `bundle plan`.
	require.NotEmpty(t, long)
	assert.NotContains(t, long, "bundle deploy")
	assert.NotContains(t, long, "bundle plan")
	// Catch any leftover "bundle"-framed phrasing anywhere in the long help.
	assert.NotContains(t, strings.ToLower(long), "bundle")
	// Sanity: it should mention the ucm flow.
	assert.Contains(t, long, "ucm deploy")
}

// TestMigrate_NoPlanCheckFlag locks in the --noplancheck flag so a future
// refactor doesn't silently drop it. The flag is the user's escape hatch
// when the spawned `ucm plan` invocation can't be reproduced cleanly
// (e.g. CI that already vetted the deploy).
func TestMigrate_NoPlanCheckFlag(t *testing.T) {
	cmd := newMigrateCommand()
	flag := cmd.Flags().Lookup("noplancheck")
	require.NotNil(t, flag, "migrate must expose --noplancheck")
	assert.Equal(t, "false", flag.DefValue)
}

// TestReadTerraformStateHeader_MissingFile asserts we surface the OS error
// (rather than synthesizing a noop) when there is no local terraform state
// to migrate from. The migrate verb relies on this distinction to print the
// "no existing local state" guidance instead of a generic parse error.
func TestReadTerraformStateHeader_MissingFile(t *testing.T) {
	_, err := readTerraformStateHeader(filepath.Join(t.TempDir(), "missing.tfstate"))
	require.Error(t, err)
}

// TestReadTerraformStateHeader_Parses verifies the lineage + serial
// extraction matches the on-disk tfstate JSON shape — the same shape
// terraform itself emits and bundle's StateDesc reads.
func TestReadTerraformStateHeader_Parses(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "terraform.tfstate")
	body := `{"version":4,"lineage":"abc-123","serial":7,"resources":[]}`
	require.NoError(t, os.WriteFile(path, []byte(body), 0o644))

	hdr, err := readTerraformStateHeader(path)
	require.NoError(t, err)
	assert.Equal(t, "abc-123", hdr.Lineage)
	assert.Equal(t, 7, hdr.Serial)
}

// TestReadTerraformStateHeader_RejectsMalformed surfaces a clear parse
// error rather than letting a corrupt tfstate slip through and produce a
// blank-lineage migrated database (which would mint a fresh lineage and
// silently break terraform-side rollback).
func TestReadTerraformStateHeader_RejectsMalformed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "terraform.tfstate")
	require.NoError(t, os.WriteFile(path, []byte("not json"), 0o644))

	_, err := readTerraformStateHeader(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse")
}

// TestMigrateAbortsWhenTerraformStateMissing exercises the "no existing
// local state" guidance path: when terraform.tfstate is absent at the
// canonical local path, migrate must print the user-facing guidance and
// return root.ErrAlreadyPrinted (so the printed message is not duplicated
// by the root error renderer).
func TestMigrateAbortsWhenTerraformStateMissing(t *testing.T) {
	u := setupUcmFixture(t)
	ctx, stderr := cmdio.NewTestContextWithStderr(t.Context())

	localTerraformPath := deploy.LocalTfStatePath(u)
	// Sanity: fixture must not pre-create the tfstate; the guard relies on
	// os.ErrNotExist as the trigger.
	_, statErr := os.Stat(localTerraformPath)
	require.True(t, errors.Is(statErr, os.ErrNotExist), "fixture leaked a terraform.tfstate at %s", localTerraformPath)

	err := checkLocalTerraformStatePresent(ctx, localTerraformPath)
	require.ErrorIs(t, err, root.ErrAlreadyPrinted)
	assert.Contains(t, stderr.String(), "no existing local state was found")
	assert.Contains(t, stderr.String(), "DATABRICKS_UCM_ENGINE=direct")
}

// TestMigrateAbortsWhenDirectStateAlreadyExists exercises the
// "resources.json already exists" guard. A user who has already migrated
// (or deployed direct from scratch) must not have their direct state
// silently overwritten — the verb errors out with a wrapped, non-printed
// error so the standard error renderer surfaces it.
func TestMigrateAbortsWhenDirectStateAlreadyExists(t *testing.T) {
	u := setupUcmFixture(t)

	localPath := phases.DirectStatePath(u)
	tempStatePath := localPath + ".temp-migration"

	require.NoError(t, os.MkdirAll(filepath.Dir(localPath), 0o755))
	require.NoError(t, os.WriteFile(localPath, []byte("{}"), 0o644))

	err := checkDirectStateAbsent(localPath, tempStatePath)
	require.Error(t, err)
	assert.NotErrorIs(t, err, root.ErrAlreadyPrinted)
	assert.Contains(t, err.Error(), "state file "+localPath+" already exists")
}
