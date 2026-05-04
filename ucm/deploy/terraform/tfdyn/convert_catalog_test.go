package tfdyn

import (
	"encoding/json"
	"testing"

	"github.com/databricks/databricks-sdk-go/service/catalog"

	"github.com/databricks/cli/libs/dyn"
	"github.com/databricks/cli/libs/dyn/convert"
	"github.com/databricks/cli/ucm/config/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertCatalog(t *testing.T) {
	tests := []struct {
		name string
		key  string
		src  resources.Catalog
		want map[string]any
	}{
		{
			name: "minimal",
			key:  "sales",
			src:  resources.Catalog{CreateCatalog: catalog.CreateCatalog{Name: "sales_prod"}},
			want: map[string]any{
				"name":          "sales_prod",
				"force_destroy": true,
			},
		},
		{
			name: "with comment and storage root",
			key:  "sales",
			src: resources.Catalog{CreateCatalog: catalog.CreateCatalog{Name: "sales_prod", Comment: "Sales team catalog", StorageRoot: "s3://bucket/root"}},
			want: map[string]any{
				"name":          "sales_prod",
				"comment":       "Sales team catalog",
				"storage_root":  "s3://bucket/root",
				"force_destroy": true,
			},
		},
		{
			name: "with tags -> properties",
			key:  "sales",
			src: resources.Catalog{CreateCatalog: catalog.CreateCatalog{Name: "sales_prod"}, Tags: map[string]string{"team": "sales", "env": "prod"}},
			want: map[string]any{
				"name": "sales_prod",
				"properties": map[string]any{
					"team": "sales",
					"env":  "prod",
				},
				"force_destroy": true,
			},
		},
		{
			name: "defaults name from key when missing",
			key:  "analytics",
			src:  resources.Catalog{},
			want: map[string]any{
				"name":          "analytics",
				"force_destroy": true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			vin, err := convert.FromTyped(tc.src, dyn.NilValue)
			require.NoError(t, err)

			out := NewResources()
			err = catalogConverter{}.Convert(t.Context(), tc.key, vin, out)
			require.NoError(t, err)

			got, ok := out.Catalog[tc.key]
			require.True(t, ok)
			assert.Equal(t, tc.want, got.AsAny())
		})
	}
}

// TestConvertCatalog_DeterministicProperties asserts that converting a
// catalog with multi-key tags produces byte-identical JSON across many
// iterations. Without explicit key-sorting the underlying Go map's
// randomized iteration leaks into the emitted properties block.
func TestConvertCatalog_DeterministicProperties(t *testing.T) {
	src := resources.Catalog{
		CreateCatalog: catalog.CreateCatalog{Name: "sales_prod"},
		Tags: map[string]string{
			"team": "sales",
			"env":  "prod",
			"cost": "alpha",
			"tier": "gold",
			"zone": "us-west",
		},
	}

	var first []byte
	for i := range 100 {
		vin, err := convert.FromTyped(src, dyn.NilValue)
		require.NoError(t, err)

		out := NewResources()
		require.NoError(t, catalogConverter{}.Convert(t.Context(), "sales", vin, out))

		buf, err := json.Marshal(out.Catalog["sales"].AsAny())
		require.NoError(t, err)

		if i == 0 {
			first = buf
			continue
		}
		require.Equal(t, string(first), string(buf), "iteration %d diverged", i)
	}
}
