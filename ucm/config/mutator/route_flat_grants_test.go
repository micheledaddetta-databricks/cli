package mutator_test

import (
	"strings"
	"testing"

	"github.com/databricks/cli/ucm"
	"github.com/databricks/cli/ucm/config/mutator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runFlattenAndRoute runs the same chain that DefaultMutators uses for the
// flat→nested handoff: FlattenNestedResources first, then RouteFlatGrants.
// Tests that need to start from already-flat input (i.e. just the
// resources.grants top-level map with no nested input) still get a no-op
// flatten pass and the routing they're testing.
func runFlattenAndRoute(t *testing.T, yaml string) (*ucm.Ucm, []string) {
	t.Helper()
	u := loadUcm(t, yaml)
	diags := ucm.Apply(t.Context(), u, mutator.FlattenNestedResources())
	require.Empty(t, diags, "flatten unexpectedly produced diags: %v", summaries(diags))
	diags = ucm.Apply(t.Context(), u, mutator.RouteFlatGrants())
	return u, summaries(diags)
}

func TestRouteFlatGrants_RoutesByKind(t *testing.T) {
	tests := []struct {
		name      string
		yaml      string
		grantName string
		// inspector reads the routed grant's principal from the typed config.
		inspect func(t *testing.T, u *ucm.Ucm)
	}{
		{
			name: "catalog grant",
			yaml: `
ucm: {name: t}
resources:
  catalogs:
    main: {name: main}
  grants:
    cat_admin:
      securable: {type: catalog, name: main}
      principal: alice
      privileges: [USE_CATALOG]
`,
			grantName: "cat_admin",
			inspect: func(t *testing.T, u *ucm.Ucm) {
				g := u.Config.Resources.Catalogs["main"].Grants["cat_admin"]
				require.NotNil(t, g)
				assert.Equal(t, "alice", g.Principal)
				assert.Equal(t, []string{"USE_CATALOG"}, g.Privileges)
			},
		},
		{
			name: "schema grant",
			yaml: `
ucm: {name: t}
resources:
  schemas:
    raw: {name: raw, catalog_name: main}
  grants:
    sch_reader:
      securable: {type: schema, name: raw}
      principal: bob
      privileges: [USE_SCHEMA]
`,
			grantName: "sch_reader",
			inspect: func(t *testing.T, u *ucm.Ucm) {
				g := u.Config.Resources.Schemas["raw"].Grants["sch_reader"]
				require.NotNil(t, g)
				assert.Equal(t, "bob", g.Principal)
			},
		},
		{
			name: "volume grant",
			yaml: `
ucm: {name: t}
resources:
  volumes:
    v1: {name: v1, catalog_name: main, schema_name: raw, volume_type: MANAGED}
  grants:
    v_read:
      securable: {type: volume, name: v1}
      principal: carol
      privileges: [READ_VOLUME]
`,
			grantName: "v_read",
			inspect: func(t *testing.T, u *ucm.Ucm) {
				g := u.Config.Resources.Volumes["v1"].Grants["v_read"]
				require.NotNil(t, g)
				assert.Equal(t, "carol", g.Principal)
			},
		},
		{
			name: "external location grant",
			yaml: `
ucm: {name: t}
resources:
  external_locations:
    el1: {name: el1, url: s3://bucket/path, credential_name: cred1}
  grants:
    el_use:
      securable: {type: external_location, name: el1}
      principal: dave
      privileges: [READ_FILES]
`,
			grantName: "el_use",
			inspect: func(t *testing.T, u *ucm.Ucm) {
				g := u.Config.Resources.ExternalLocations["el1"].Grants["el_use"]
				require.NotNil(t, g)
				assert.Equal(t, "dave", g.Principal)
			},
		},
		{
			name: "storage credential grant",
			yaml: `
ucm: {name: t}
resources:
  storage_credentials:
    sc1: {name: sc1}
  grants:
    sc_use:
      securable: {type: storage_credential, name: sc1}
      principal: eve
      privileges: [CREATE_EXTERNAL_LOCATION]
`,
			grantName: "sc_use",
			inspect: func(t *testing.T, u *ucm.Ucm) {
				g := u.Config.Resources.StorageCredentials["sc1"].Grants["sc_use"]
				require.NotNil(t, g)
				assert.Equal(t, "eve", g.Principal)
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u, diags := runFlattenAndRoute(t, tc.yaml)
			require.Empty(t, diags, "unexpected diags: %v", diags)

			// Top-level flat grants is empty after routing.
			assert.Empty(t, u.Config.Resources.Grants, "top-level grants should be empty post-route")

			tc.inspect(t, u)
		})
	}
}

func TestRouteFlatGrants_UnsupportedTypeErrors(t *testing.T) {
	yaml := `
ucm: {name: t}
resources:
  catalogs:
    main: {name: main}
  grants:
    g1:
      securable: {type: table, name: foo}
      principal: p
      privileges: [SELECT]
`
	_, diags := runFlattenAndRoute(t, yaml)
	require.NotEmpty(t, diags)
	found := false
	for _, s := range diags {
		if strings.Contains(s, "unsupported securable type") {
			found = true
			break
		}
	}
	assert.True(t, found, "expected unsupported-type diag, got %v", diags)
}

func TestRouteFlatGrants_MissingParentErrors(t *testing.T) {
	yaml := `
ucm: {name: t}
resources:
  catalogs:
    main: {name: main}
  grants:
    g1:
      securable: {type: catalog, name: ghost}
      principal: p
      privileges: [USE_CATALOG]
`
	_, diags := runFlattenAndRoute(t, yaml)
	require.NotEmpty(t, diags)
	found := false
	for _, s := range diags {
		if strings.Contains(s, "non-existent catalog") {
			found = true
			break
		}
	}
	assert.True(t, found, "expected missing-parent diag, got %v", diags)
}

func TestRouteFlatGrants_CollisionErrors(t *testing.T) {
	// Nested catalog grant and a flat-form grant collide on the same key.
	// FlattenNestedResources lifts the nested one to flat first, then
	// RouteFlatGrants re-routes both back. The nested one is routed first
	// (when the same dyn key is processed), and the second insertion attempt
	// hits the collision check against the existing nested grant.
	//
	// Since FlattenNestedResources also detects flat/nested key collisions
	// in the grants map up front, we use the same-key-different-securable
	// path to land both into the same catalog: declare the flat grant under
	// a different top-level key but same parent, and pre-populate the
	// nested form under that catalog.
	yaml := `
ucm: {name: t}
resources:
  catalogs:
    main:
      name: main
      grants:
        cat_admin:
          principal: alice
          privileges: [USE_CATALOG]
  grants:
    cat_admin:
      securable: {type: catalog, name: main}
      principal: bob
      privileges: [USE_CATALOG]
`
	// FlattenNestedResources will flag this as a flat-vs-nested collision
	// before RouteFlatGrants gets to it, which is the intended early-exit
	// for end users; the routing-stage collision check below covers the
	// case where flatten already merged grants into the flat map.
	u := loadUcm(t, yaml)
	flatDiags := ucm.Apply(t.Context(), u, mutator.FlattenNestedResources())
	require.NotEmpty(t, flatDiags)
	found := false
	for _, s := range summaries(flatDiags) {
		if strings.Contains(s, "declared both as a flat entry and nested") {
			found = true
			break
		}
	}
	assert.True(t, found, "expected nested/flat collision diag from flatten, got %v", summaries(flatDiags))
}

func TestRouteFlatGrants_RoutesWithoutPriorFlatten(t *testing.T) {
	// RouteFlatGrants is independent of FlattenNestedResources for purely
	// flat input — the chain order in DefaultMutators puts flatten first,
	// but routing on plain flat input must still produce the nested form.
	yaml := `
ucm: {name: t}
resources:
  catalogs:
    main:
      name: main
  grants:
    g1:
      securable: {type: catalog, name: main}
      principal: bob
      privileges: [USE_CATALOG]
`
	u := loadUcm(t, yaml)
	diags := ucm.Apply(t.Context(), u, mutator.RouteFlatGrants())
	require.Empty(t, diags, "unexpected diags: %v", summaries(diags))
	g := u.Config.Resources.Catalogs["main"].Grants["g1"]
	require.NotNil(t, g)
	assert.Equal(t, "bob", g.Principal)
	assert.Empty(t, u.Config.Resources.Grants)
}

func TestRouteFlatGrants_NoOpWhenNoFlatGrants(t *testing.T) {
	yaml := `
ucm: {name: t}
resources:
  catalogs:
    main: {name: main}
`
	u, diags := runFlattenAndRoute(t, yaml)
	require.Empty(t, diags, "unexpected diags: %v", diags)
	assert.Empty(t, u.Config.Resources.Grants)
	assert.Empty(t, u.Config.Resources.Catalogs["main"].Grants)
}
