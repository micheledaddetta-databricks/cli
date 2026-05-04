package tfdyn

import (
	"context"
	"fmt"
	"sort"

	"github.com/databricks/cli/libs/dyn"
	"github.com/databricks/cli/ucm"
)

// providerSource and providerVersion mirror the constants in the parent
// terraform package. Duplicated (rather than imported) to keep this
// subpackage import-cycle-free — the parent package imports tfdyn.
const (
	providerSource  = "databricks/databricks"
	providerVersion = "1.112.0"
)

// convertOrder controls the order in which resource kinds are walked. The
// ordering matters because downstream converters inspect earlier ones
// (schemas look at Resources.Catalog to decide whether to emit depends_on;
// grants look at Resources.Catalog and Resources.Schema).
var convertOrder = []string{"storage_credentials", "external_locations", "catalogs", "schemas", "volumes", "connections", "grants"}

// grantHostKinds enumerates the per-resource buckets whose `.grants` child
// map is harvested into a synthetic flat-grants view at dispatch time. Mirrors
// the kinds that mutator.RouteFlatGrants targets — keep in sync if a new
// grantable kind is added.
var grantHostKinds = []string{"catalogs", "schemas", "volumes", "external_locations", "storage_credentials"}

// Convert walks a ucm configuration and produces the Terraform JSON
// resource tree suitable for writing as a .tf.json file. The returned
// dyn.Value is shaped as:
//
//	{
//	  "resource": {
//	    "databricks_catalog": { "<key>": { ... } },
//	    "databricks_schema":  { "<key>": { ... } },
//	    "databricks_grants":  { "<key>": { ... } }
//	  }
//	}
//
// Empty resource kinds are omitted.
func Convert(ctx context.Context, u *ucm.Ucm) (dyn.Value, error) {
	out := NewResources()

	resourcesVal, err := dyn.GetByPath(u.Config.Value(), dyn.NewPath(dyn.Key("resources")))
	if err != nil {
		// No resources: emit an empty terraform file rather than failing.
		resourcesVal = dyn.V(map[string]dyn.Value{})
	}

	resourcesVal, err = liftNestedGrantsForDispatch(resourcesVal)
	if err != nil {
		return dyn.InvalidValue, err
	}

	for _, kind := range convertOrder {
		conv, ok := GetConverter(kind)
		if !ok {
			continue
		}
		bucket := resourcesVal.Get(kind)
		if !bucket.IsValid() {
			continue
		}
		m, ok := bucket.AsMap()
		if !ok {
			continue
		}

		keys := make([]string, 0, m.Len())
		for _, p := range m.Pairs() {
			keys = append(keys, p.Key.MustString())
		}
		sort.Strings(keys)

		for _, key := range keys {
			vin, _ := m.GetByString(key)
			if err := conv.Convert(ctx, key, vin, out); err != nil {
				return dyn.InvalidValue, fmt.Errorf("convert %s.%s: %w", kind, key, err)
			}
		}
	}

	return buildResourceTree(out), nil
}

// liftNestedGrantsForDispatch walks the per-resource `grants` maps and
// merges them into a synthetic top-level `grants` map so the existing
// `grants` converter can process every grant in one pass. mutator.RouteFlatGrants
// has already moved any flat-form entries into the nested form by the time
// this runs in production, leaving `resources.grants` empty; lifting the
// nested form back to a flat shape here keeps the dispatch contract stable
// without registering five additional kind-specific converters.
//
// Each nested grant body carries `securable` populated by the mutator (or by
// the user, for nested-form grants that bypass routing); the converter reads
// it directly. Dyn locations are preserved end-to-end so diagnostics still
// point at the originating ucm.yml span.
func liftNestedGrantsForDispatch(resourcesVal dyn.Value) (dyn.Value, error) {
	resources, ok := resourcesVal.AsMap()
	if !ok {
		return resourcesVal, nil
	}

	flatGrants := mapOrNew(getByString(resources, "grants"))

	for _, kind := range grantHostKinds {
		bucketVal, _ := resources.GetByString(kind)
		bucket, ok := bucketVal.AsMap()
		if !ok {
			continue
		}
		for _, parent := range bucket.Pairs() {
			parentMap, ok := parent.Value.AsMap()
			if !ok {
				continue
			}
			grantsVal, _ := parentMap.GetByString("grants")
			grants, ok := grantsVal.AsMap()
			if !ok {
				continue
			}
			for _, gp := range grants.Pairs() {
				key := gp.Key.MustString()
				if _, exists := flatGrants.GetByString(key); exists {
					return dyn.InvalidValue, fmt.Errorf("grant %q: nested entry under %s.%s collides with flat entry", key, kind, parent.Key.MustString())
				}
				flatGrants.SetLoc(key, gp.Key.Locations(), gp.Value)
			}
		}
	}

	if flatGrants.Len() == 0 {
		return resourcesVal, nil
	}

	newResources := resources.Clone()
	existingFlat := getByString(resources, "grants")
	newResources.SetLoc("grants", existingFlat.Locations(),
		dyn.NewValue(flatGrants, existingFlat.Locations()))
	return dyn.NewValue(newResources, resourcesVal.Locations()), nil
}

// getByString fetches a child value from a dyn.Mapping, returning the zero
// dyn.Value if the key is absent. Mirrors the small helper pattern used in
// the mutator package without taking a cross-package dependency.
func getByString(m dyn.Mapping, key string) dyn.Value {
	v, _ := m.GetByString(key)
	return v
}

// mapOrNew returns the underlying mapping of v cloned, or an empty mapping
// if v is not a map. Local copy of the helper from the mutator package; the
// dispatch only needs the same fall-through semantics.
func mapOrNew(v dyn.Value) dyn.Mapping {
	m, ok := v.AsMap()
	if !ok {
		return dyn.NewMapping()
	}
	return m.Clone()
}

// buildResourceTree assembles the top-level Terraform JSON tree. The output
// shape mirrors what bundle/deploy/terraform writes so `terraform init`
// resolves the databricks provider out of the databricks/databricks
// namespace instead of defaulting to hashicorp/databricks (which does not
// exist in the registry).
func buildResourceTree(out *Resources) dyn.Value {
	blocks := []struct {
		tfType string
		values map[string]dyn.Value
	}{
		{"databricks_storage_credential", out.StorageCredential},
		{"databricks_external_location", out.ExternalLocation},
		{"databricks_catalog", out.Catalog},
		{"databricks_schema", out.Schema},
		{"databricks_volume", out.Volume},
		{"databricks_connection", out.Connection},
		{"databricks_grants", out.Grants},
	}

	var resourcePairs []dyn.Pair
	for _, b := range blocks {
		if len(b.values) == 0 {
			continue
		}
		keys := make([]string, 0, len(b.values))
		for k := range b.values {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		pairs := make([]dyn.Pair, 0, len(keys))
		for _, k := range keys {
			pairs = append(pairs, dyn.Pair{
				Key:   dyn.V(k),
				Value: b.values[k],
			})
		}
		resourcePairs = append(resourcePairs, dyn.Pair{
			Key:   dyn.V(b.tfType),
			Value: dyn.V(dyn.NewMappingFromPairs(pairs)),
		})
	}

	rootPairs := []dyn.Pair{
		{Key: dyn.V("terraform"), Value: buildTerraformBlock()},
		{Key: dyn.V("provider"), Value: buildProviderBlock()},
		{
			Key:   dyn.V("resource"),
			Value: dyn.V(dyn.NewMappingFromPairs(resourcePairs)),
		},
	}
	return dyn.V(dyn.NewMappingFromPairs(rootPairs))
}

// buildTerraformBlock returns the `terraform.required_providers.databricks`
// value that pins the provider source and version.
func buildTerraformBlock() dyn.Value {
	databricks := dyn.V(dyn.NewMappingFromPairs([]dyn.Pair{
		{Key: dyn.V("source"), Value: dyn.V(providerSource)},
		{Key: dyn.V("version"), Value: dyn.V(providerVersion)},
	}))
	required := dyn.V(dyn.NewMappingFromPairs([]dyn.Pair{
		{Key: dyn.V("databricks"), Value: databricks},
	}))
	return dyn.V(dyn.NewMappingFromPairs([]dyn.Pair{
		{Key: dyn.V("required_providers"), Value: required},
	}))
}

// buildProviderBlock returns an empty `provider.databricks` block. The
// databricks terraform provider picks up its auth from the DATABRICKS_*
// env vars that buildEnv passes through to terraform-exec, so no fields
// need to be set here — the block's presence is what matters.
func buildProviderBlock() dyn.Value {
	return dyn.V(dyn.NewMappingFromPairs([]dyn.Pair{
		{Key: dyn.V("databricks"), Value: dyn.V(dyn.NewMapping())},
	}))
}
