package ucm

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/databricks/cli/ucm/deploy"
	"github.com/databricks/cli/ucm/direct/dstate"
	"github.com/databricks/cli/ucm/phases"
	"github.com/databricks/databricks-sdk-go/service/catalog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedDirectStateCatalog writes a v2 direct-engine state file under
// workDir/.databricks/ucm/<target>/ so the drift command's dstate.Database.Open
// picks it up. Mirrors the on-disk shape produced by ucm/direct.Apply on a
// successful create.
func seedDirectStateCatalog(t *testing.T, workDir, target, name, comment string) {
	t.Helper()
	statePath := filepath.Join(workDir, filepath.FromSlash(deploy.LocalCacheDir), target, "resources.json")
	var db dstate.DeploymentState
	require.NoError(t, db.Open(statePath))
	require.NoError(t, db.SaveState(
		"resources.catalogs."+name,
		name,
		&catalog.CreateCatalog{Name: name, Comment: comment},
		nil,
	))
	require.NoError(t, db.Finalize())
}

func TestCmd_Drift_NoStatePrintsNoDrift(t *testing.T) {
	_ = newVerbHarness(t)

	stdout, stderr, err := runVerb(t, validFixtureDir(t), "drift")
	t.Logf("stdout=%q stderr=%q", stdout, stderr)

	require.NoError(t, err)
	assert.Contains(t, stdout, "No drift detected")
}

// TestRenderDriftText_NoDriftPrintsPositiveLine exercises the renderer
// directly — the integration tests cover the full verb path, and renderer
// tests assert the exact wire format without depending on the cobra harness
// wiring the -o flag (only root.New does that, not New() in isolation).
func TestRenderDriftText_NoDriftPrintsPositiveLine(t *testing.T) {
	var buf bytes.Buffer
	renderDriftText(&buf, &phases.Report{})
	assert.Equal(t, "No drift detected.\n", buf.String())
}

func TestRenderDriftText_DriftPrintsSpecFormat(t *testing.T) {
	var buf bytes.Buffer
	report := &phases.Report{Drift: []phases.ResourceDrift{
		{
			Key:    "resources.catalogs.sales",
			Fields: []phases.FieldDrift{{Field: "comment", State: "sales data", Live: "sales domain data"}},
		},
		{
			Key:    "resources.external_locations.shared",
			Fields: []phases.FieldDrift{{Field: "read_only", State: false, Live: true}},
		},
	}}
	renderDriftText(&buf, report)
	got := buf.String()
	assert.Contains(t, got, "DRIFT DETECTED on 2 resource(s):")
	assert.Contains(t, got, `  comment: state="sales data" live="sales domain data"`)
	assert.Contains(t, got, `  read_only: state=false live=true`)
}

func TestRenderDriftJSON_ProducesDriftKey(t *testing.T) {
	var buf bytes.Buffer
	report := &phases.Report{Drift: []phases.ResourceDrift{
		{
			Key:    "resources.catalogs.sales",
			Fields: []phases.FieldDrift{{Field: "comment", State: "a", Live: "b"}},
		},
	}}
	require.NoError(t, renderDriftJSON(&buf, report))

	var round struct {
		Drift []struct {
			Key    string `json:"key"`
			Fields []struct {
				Field string `json:"field"`
				State any    `json:"state"`
				Live  any    `json:"live"`
			} `json:"fields"`
		} `json:"drift"`
	}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &round))
	require.Len(t, round.Drift, 1)
	assert.Equal(t, "resources.catalogs.sales", round.Drift[0].Key)
	require.Len(t, round.Drift[0].Fields, 1)
	assert.Equal(t, "comment", round.Drift[0].Fields[0].Field)
	assert.Equal(t, "a", round.Drift[0].Fields[0].State)
	assert.Equal(t, "b", round.Drift[0].Fields[0].Live)
}
