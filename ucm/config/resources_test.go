package config_test

import (
	"testing"

	"github.com/databricks/cli/ucm/config"
	"github.com/databricks/cli/ucm/config/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResources_AllGrants_AcrossAllKinds(t *testing.T) {
	r := &config.Resources{
		Catalogs: map[string]*resources.Catalog{
			"cat1": {Grants: map[string]*resources.Grant{
				"g_cat": {Principal: "alice", Privileges: []string{"USE_CATALOG"}},
			}},
		},
		Schemas: map[string]*resources.Schema{
			"sch1": {Grants: map[string]*resources.Grant{
				"g_sch": {Principal: "bob", Privileges: []string{"USE_SCHEMA"}},
			}},
		},
		Volumes: map[string]*resources.Volume{
			"vol1": {Grants: map[string]*resources.Grant{
				"g_vol": {Principal: "carol", Privileges: []string{"READ_VOLUME"}},
			}},
		},
		ExternalLocations: map[string]*resources.ExternalLocation{
			"el1": {Grants: map[string]*resources.Grant{
				"g_el": {Principal: "dave", Privileges: []string{"READ_FILES"}},
			}},
		},
		StorageCredentials: map[string]*resources.StorageCredential{
			"sc1": {Grants: map[string]*resources.Grant{
				"g_sc": {Principal: "eve", Privileges: []string{"CREATE_EXTERNAL_LOCATION"}},
			}},
		},
	}

	got := r.AllGrants()
	require.Len(t, got, 5)

	assert.Equal(t, "catalog", got["g_cat"].Securable.Type)
	assert.Equal(t, "cat1", got["g_cat"].Securable.Name)
	assert.Equal(t, "schema", got["g_sch"].Securable.Type)
	assert.Equal(t, "sch1", got["g_sch"].Securable.Name)
	assert.Equal(t, "volume", got["g_vol"].Securable.Type)
	assert.Equal(t, "vol1", got["g_vol"].Securable.Name)
	assert.Equal(t, "external_location", got["g_el"].Securable.Type)
	assert.Equal(t, "el1", got["g_el"].Securable.Name)
	assert.Equal(t, "storage_credential", got["g_sc"].Securable.Type)
	assert.Equal(t, "sc1", got["g_sc"].Securable.Name)
}

func TestResources_AllGrants_Empty(t *testing.T) {
	r := &config.Resources{}
	assert.Empty(t, r.AllGrants())
}

func TestResources_AllGrants_PreservesExistingSecurable(t *testing.T) {
	r := &config.Resources{
		Catalogs: map[string]*resources.Catalog{
			"cat1": {Grants: map[string]*resources.Grant{
				"g": {
					Securable:  resources.Securable{Type: "catalog", Name: "cat1"},
					Principal:  "alice",
					Privileges: []string{"USE_CATALOG"},
				},
			}},
		},
	}

	got := r.AllGrants()
	require.Contains(t, got, "g")
	assert.Equal(t, "catalog", got["g"].Securable.Type)
	assert.Equal(t, "cat1", got["g"].Securable.Name)
}

func TestResources_AllGrants_SkipsNilEntries(t *testing.T) {
	r := &config.Resources{
		Catalogs: map[string]*resources.Catalog{
			"cat1": {Grants: map[string]*resources.Grant{
				"g_nil": nil,
				"g_ok": {
					Principal:  "alice",
					Privileges: []string{"USE_CATALOG"},
				},
			}},
		},
	}

	got := r.AllGrants()
	require.Len(t, got, 1)
	assert.Contains(t, got, "g_ok")
}
