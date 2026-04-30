package phases_test

import (
	"errors"
	"testing"

	"github.com/databricks/cli/libs/logdiag"
	"github.com/databricks/cli/ucm/direct/dstate"
	"github.com/databricks/cli/ucm/phases"
	"github.com/databricks/databricks-sdk-go/apierr"
	"github.com/databricks/databricks-sdk-go/service/catalog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDriftReturnsEmptyReportOnMissingState(t *testing.T) {
	f := newFixture(t)
	ctx := logdiag.InitContext(t.Context())
	logdiag.SetCollect(ctx, true)

	report := phases.Drift(ctx, f.u, phases.Options{})

	require.False(t, logdiag.HasError(ctx), "unexpected errors: %v", logdiag.FlushCollected(ctx))
	require.NotNil(t, report)
	assert.False(t, report.HasDrift())
}

func TestDriftReportsFieldMismatch(t *testing.T) {
	f := newFixture(t)
	seedDirectStateCatalog(t, phases.DirectStatePath(f.u), "sales")

	// State carries a default-empty catalog; live read returns a mismatching comment.
	f.mockWS.GetMockCatalogsAPI().EXPECT().
		GetByName(mock.Anything, "sales").
		Return(&catalog.CatalogInfo{Name: "sales", Comment: "drifted"}, nil)

	ctx := logdiag.InitContext(t.Context())
	logdiag.SetCollect(ctx, true)
	report := phases.Drift(ctx, f.u, phases.Options{})

	require.False(t, logdiag.HasError(ctx), "unexpected errors: %v", logdiag.FlushCollected(ctx))
	require.NotNil(t, report)
	require.True(t, report.HasDrift())
	require.Len(t, report.Drift, 1)
	assert.Equal(t, "resources.catalogs.sales", report.Drift[0].Key)
}

func TestDriftRecordsExistsFalseWhenLiveGone(t *testing.T) {
	f := newFixture(t)
	seedDirectStateCatalog(t, phases.DirectStatePath(f.u), "gone")

	f.mockWS.GetMockCatalogsAPI().EXPECT().
		GetByName(mock.Anything, "gone").
		Return(nil, apierr.ErrResourceDoesNotExist)

	ctx := logdiag.InitContext(t.Context())
	logdiag.SetCollect(ctx, true)
	report := phases.Drift(ctx, f.u, phases.Options{})

	require.False(t, logdiag.HasError(ctx), "unexpected errors: %v", logdiag.FlushCollected(ctx))
	require.NotNil(t, report)
	require.True(t, report.HasDrift())
	require.Len(t, report.Drift, 1)
	require.Len(t, report.Drift[0].Fields, 1)
	assert.Equal(t, "_exists", report.Drift[0].Fields[0].Field)
	assert.Equal(t, true, report.Drift[0].Fields[0].State)
	assert.Equal(t, false, report.Drift[0].Fields[0].Live)
}

func TestDriftBailsOnSDKError(t *testing.T) {
	f := newFixture(t)
	seedDirectStateCatalog(t, phases.DirectStatePath(f.u), "sales")

	f.mockWS.GetMockCatalogsAPI().EXPECT().
		GetByName(mock.Anything, "sales").
		Return(nil, errors.New("500 internal"))

	ctx := logdiag.InitContext(t.Context())
	logdiag.SetCollect(ctx, true)
	report := phases.Drift(ctx, f.u, phases.Options{})

	assert.Nil(t, report)
	require.True(t, logdiag.HasError(ctx))
}

func TestDriftSkipsGrantsEntries(t *testing.T) {
	f := newFixture(t)

	// Seed both a non-grants and a grants entry; only the non-grants entry
	// should produce a Get* call. The grants entry must be silently ignored.
	statePath := phases.DirectStatePath(f.u)
	var db dstate.DeploymentState
	require.NoError(t, db.Open(statePath))
	require.NoError(t, db.SaveState("resources.catalogs.sales", "sales", &catalog.CreateCatalog{Name: "sales"}, nil))
	require.NoError(t, db.SaveState("resources.catalogs.sales.grants", "sales", []catalog.PrivilegeAssignment{}, nil))
	require.NoError(t, db.Finalize())

	f.mockWS.GetMockCatalogsAPI().EXPECT().
		GetByName(mock.Anything, "sales").
		Return(&catalog.CatalogInfo{Name: "sales"}, nil)

	ctx := logdiag.InitContext(t.Context())
	logdiag.SetCollect(ctx, true)
	report := phases.Drift(ctx, f.u, phases.Options{})

	require.False(t, logdiag.HasError(ctx), "unexpected errors: %v", logdiag.FlushCollected(ctx))
	require.NotNil(t, report)
	assert.False(t, report.HasDrift())
}
