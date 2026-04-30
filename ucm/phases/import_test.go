package phases_test

import (
	"errors"
	"os"
	"testing"

	"github.com/databricks/cli/libs/logdiag"
	"github.com/databricks/cli/ucm/config/engine"
	"github.com/databricks/cli/ucm/config/resources"
	"github.com/databricks/cli/ucm/direct/dstate"
	"github.com/databricks/cli/ucm/phases"
	"github.com/databricks/databricks-sdk-go/apierr"
	"github.com/databricks/databricks-sdk-go/service/catalog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestImportTerraformEngineRunsImportAndPushes(t *testing.T) {
	f := newFixture(t)
	f.u.Config.Resources.Catalogs = map[string]*resources.Catalog{
		"main": {CreateCatalog: catalog.CreateCatalog{Name: "main"}},
	}
	ctx := logdiag.InitContext(t.Context())
	logdiag.SetCollect(ctx, true)

	phases.Import(ctx, f.u, phases.Options{
		Backend:          f.backend,
		TerraformFactory: fakeTfFactory(f.tf),
	}, phases.ImportRequest{Kind: phases.ImportCatalog, Name: "main", Key: "main"})

	require.False(t, logdiag.HasError(ctx), "unexpected errors: %v", logdiag.FlushCollected(ctx))
	assert.Equal(t, 1, f.tf.ImportCalls)
	assert.Equal(t, "databricks_catalog.main", f.tf.LastImportAddress)
	assert.Equal(t, "main", f.tf.LastImportId)
	assert.Equal(t, 1, readRemoteSeq(t, f), "successful import must push remote state")
}

func TestImportRequiresDeclaredResource(t *testing.T) {
	f := newFixture(t)
	ctx := logdiag.InitContext(t.Context())
	logdiag.SetCollect(ctx, true)

	phases.Import(ctx, f.u, phases.Options{
		Backend:          f.backend,
		TerraformFactory: fakeTfFactory(f.tf),
	}, phases.ImportRequest{Kind: phases.ImportCatalog, Name: "ghost", Key: "ghost"})

	require.True(t, logdiag.HasError(ctx))
	assert.Equal(t, 0, f.tf.ImportCalls)
}

// TestImportDirectEnginePersistsState exercises the new direct-engine import
// flow end-to-end: GetByName via mock, RemapState into the catalog state shape,
// SaveState under resources.catalogs.<key>, Finalize writes the state file.
func TestImportDirectEnginePersistsState(t *testing.T) {
	f := newFixture(t)
	f.u.Config.Ucm.Engine = engine.EngineDirect
	f.u.Config.Resources.Catalogs = map[string]*resources.Catalog{
		"main": {CreateCatalog: catalog.CreateCatalog{Name: "main"}},
	}
	f.mockWS.GetMockCatalogsAPI().EXPECT().
		GetByName(mock.Anything, "main").
		Return(&catalog.CatalogInfo{Name: "main", Comment: "imported"}, nil)

	ctx := logdiag.InitContext(t.Context())
	logdiag.SetCollect(ctx, true)

	phases.Import(ctx, f.u, phases.Options{
		TerraformFactory: fakeTfFactory(f.tf),
	}, phases.ImportRequest{Kind: phases.ImportCatalog, Name: "main", Key: "main"})

	require.False(t, logdiag.HasError(ctx), "unexpected errors: %v", logdiag.FlushCollected(ctx))
	assert.Equal(t, 0, f.tf.ImportCalls, "direct engine must not invoke the terraform wrapper")
	assert.Equal(t, -1, readRemoteSeq(t, f), "direct engine must never push remote state")

	// Finalize wrote the state file with the imported catalog entry.
	statePath := phases.DirectStatePath(f.u)
	info, err := os.Stat(statePath)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))

	var db dstate.DeploymentState
	require.NoError(t, db.Open(statePath))
	entry, ok := db.Data.State["resources.catalogs.main"]
	require.True(t, ok, "expected resources.catalogs.main entry after import")
	assert.Equal(t, "main", entry.ID)
}

// TestImportDirectEngineRefusesAlreadyBound asserts the pre-DoRead state
// lookup short-circuits when the key is already in state, so re-imports
// surface a clean error instead of silently overwriting recorded fields.
func TestImportDirectEngineRefusesAlreadyBound(t *testing.T) {
	f := newFixture(t)
	f.u.Config.Ucm.Engine = engine.EngineDirect
	f.u.Config.Resources.Catalogs = map[string]*resources.Catalog{
		"main": {CreateCatalog: catalog.CreateCatalog{Name: "main"}},
	}
	seedDirectStateCatalog(t, phases.DirectStatePath(f.u), "main")

	ctx := logdiag.InitContext(t.Context())
	logdiag.SetCollect(ctx, true)

	phases.Import(ctx, f.u, phases.Options{
		TerraformFactory: fakeTfFactory(f.tf),
	}, phases.ImportRequest{Kind: phases.ImportCatalog, Name: "main", Key: "main"})

	require.True(t, logdiag.HasError(ctx))
	diags := logdiag.FlushCollected(ctx)
	require.NotEmpty(t, diags)
	assert.Contains(t, diags[0].Summary, "already bound in state")
}

// TestImportDirectEngineSurfacesNotFound asserts a 404 from the SDK is
// translated into a friendly diagnostic instead of being wrapped raw.
func TestImportDirectEngineSurfacesNotFound(t *testing.T) {
	f := newFixture(t)
	f.u.Config.Ucm.Engine = engine.EngineDirect
	f.u.Config.Resources.Catalogs = map[string]*resources.Catalog{
		"missing": {CreateCatalog: catalog.CreateCatalog{Name: "missing"}},
	}
	f.mockWS.GetMockCatalogsAPI().EXPECT().
		GetByName(mock.Anything, "missing").
		Return(nil, apierr.ErrResourceDoesNotExist)

	ctx := logdiag.InitContext(t.Context())
	logdiag.SetCollect(ctx, true)

	phases.Import(ctx, f.u, phases.Options{
		TerraformFactory: fakeTfFactory(f.tf),
	}, phases.ImportRequest{Kind: phases.ImportCatalog, Name: "missing", Key: "missing"})

	require.True(t, logdiag.HasError(ctx))
	diags := logdiag.FlushCollected(ctx)
	require.NotEmpty(t, diags)
	assert.Contains(t, diags[0].Summary, "not found in Unity Catalog")

	// Finalize must not have run — no state file created.
	_, err := os.Stat(phases.DirectStatePath(f.u))
	assert.True(t, errors.Is(err, os.ErrNotExist))
}
