package mutator

import (
	"context"
	"fmt"

	"github.com/databricks/cli/libs/diag"
	"github.com/databricks/cli/libs/dyn"
	"github.com/databricks/cli/ucm"
)

// securableTypeToPlural maps a UC securable type token to the resource map's
// plural key under resources.<plural>. UCM today manages exactly these five
// kinds; any other type is rejected with a diagnostic.
var securableTypeToPlural = map[string]string{
	"catalog":            "catalogs",
	"schema":             "schemas",
	"volume":             "volumes",
	"external_location":  "external_locations",
	"storage_credential": "storage_credentials",
}

type routeFlatGrants struct{}

// RouteFlatGrants moves entries from the top-level resources.grants map
// back into the per-resource resources.<plural>.<name>.grants map indicated
// by each grant's securable.{type, name}. Designed to run after
// FlattenNestedResources so the union of nested- and flat-form grants ends
// up nested. After the mutator runs, resources.grants is empty and the
// direct engine's dresources package — which only registers nested-form
// grant adapters — sees every grant.
//
// Errors: missing/empty securable, securable.type outside the routing
// table, securable.name with no matching parent resource, or a key
// collision against an already-nested grant.
//
// NOT YET WIRED INTO DefaultMutators. Several existing consumers
// (ucm/render/groups.go grant summary, ucm/config/validate's grant checks,
// ucm/deploy/direct + ucm/deploy/terraform/tfdyn grant rendering, and
// cmd/ucm/deployment/bind_resource.go) still read the flat
// resources.grants map. Wiring this mutator now would silently zero those
// surfaces. Wiring lands in a follow-up that migrates each consumer to
// the nested form.
func RouteFlatGrants() ucm.Mutator { return &routeFlatGrants{} }

func (m *routeFlatGrants) Name() string { return "RouteFlatGrants" }

func (m *routeFlatGrants) Apply(_ context.Context, u *ucm.Ucm) diag.Diagnostics {
	var diags diag.Diagnostics
	err := u.Config.Mutate(func(root dyn.Value) (dyn.Value, error) {
		resourcesValue := root.Get("resources")
		resources, ok := resourcesValue.AsMap()
		if !ok {
			return root, nil
		}

		flatGrantsValue, _ := resources.GetByString("grants")
		flatGrants, ok := flatGrantsValue.AsMap()
		if !ok || flatGrants.Len() == 0 {
			return root, nil
		}

		// Per-plural staging: collect routed grants in maps keyed by
		// resource name so we apply a single rewrite per parent at the end.
		type stagedGrant struct {
			body dyn.Value
			key  dyn.Value
		}
		staged := make(map[string]map[string]map[string]stagedGrant)

		for _, gp := range flatGrants.Pairs() {
			grantKey := gp.Key.MustString()
			grantBody, ok := gp.Value.AsMap()
			if !ok {
				continue
			}

			secVal, secOK := grantBody.GetByString("securable")
			if !secOK || !secVal.IsValid() {
				diags = append(diags, diag.Diagnostic{
					Severity:  diag.Error,
					Summary:   fmt.Sprintf("grant %q is missing securable", grantKey),
					Locations: gp.Key.Locations(),
				})
				continue
			}
			secMap, _ := secVal.AsMap()
			typeVal, _ := secMap.GetByString("type")
			nameVal, _ := secMap.GetByString("name")
			secType, _ := typeVal.AsString()
			secName, _ := nameVal.AsString()
			if secType == "" || secName == "" {
				diags = append(diags, diag.Diagnostic{
					Severity:  diag.Error,
					Summary:   fmt.Sprintf("grant %q has empty securable.type or securable.name", grantKey),
					Locations: gp.Key.Locations(),
				})
				continue
			}

			plural, ok := securableTypeToPlural[secType]
			if !ok {
				diags = append(diags, diag.Diagnostic{
					Severity:  diag.Error,
					Summary:   fmt.Sprintf("unsupported securable type %q for grants routing", secType),
					Locations: gp.Key.Locations(),
				})
				continue
			}

			parentBucketValue, _ := resources.GetByString(plural)
			parentBucket, ok := parentBucketValue.AsMap()
			if !ok {
				diags = append(diags, diag.Diagnostic{
					Severity:  diag.Error,
					Summary:   fmt.Sprintf("grant %q references non-existent %s %q", grantKey, secType, secName),
					Locations: gp.Key.Locations(),
				})
				continue
			}
			parentValue, parentOK := parentBucket.GetByString(secName)
			if !parentOK {
				diags = append(diags, diag.Diagnostic{
					Severity:  diag.Error,
					Summary:   fmt.Sprintf("grant %q references non-existent %s %q", grantKey, secType, secName),
					Locations: gp.Key.Locations(),
				})
				continue
			}
			parentMap, ok := parentValue.AsMap()
			if !ok {
				continue
			}

			// Reject collisions with an already-nested grant of the same key.
			if existingGrants, ok := parentMap.GetByString("grants"); ok {
				if existingMap, ok := existingGrants.AsMap(); ok {
					if existing, ok := existingMap.GetPairByString(grantKey); ok {
						diags = append(diags, collisionDiag("grant", grantKey, gp.Key.Locations(), existing.Key.Locations())...)
						continue
					}
				}
			}

			body := removeKeys(grantBody, "securable")
			if staged[plural] == nil {
				staged[plural] = make(map[string]map[string]stagedGrant)
			}
			if staged[plural][secName] == nil {
				staged[plural][secName] = make(map[string]stagedGrant)
			}
			staged[plural][secName][grantKey] = stagedGrant{
				body: dyn.NewValue(body, gp.Value.Locations()),
				key:  gp.Key,
			}
		}

		newResources := resources.Clone()
		for plural, byName := range staged {
			if len(byName) == 0 {
				continue
			}
			parentBucketValue, _ := resources.GetByString(plural)
			parentBucket := mapOrNew(parentBucketValue)

			for parentName, grants := range byName {
				parentValue, _ := parentBucket.GetByString(parentName)
				parentMap := mapOrNew(parentValue)

				existingGrantsValue, _ := parentMap.GetByString("grants")
				newGrants := mapOrNew(existingGrantsValue)
				for grantKey, sg := range grants {
					newGrants.SetLoc(grantKey, sg.key.Locations(), sg.body)
				}
				parentMap.SetLoc("grants", existingGrantsValue.Locations(),
					dyn.NewValue(newGrants, existingGrantsValue.Locations()))
				parentBucket.SetLoc(parentName, nil,
					dyn.NewValue(parentMap, parentValue.Locations()))
			}
			newResources.SetLoc(plural, nil,
				dyn.NewValue(parentBucket, parentBucketValue.Locations()))
		}

		// Drop resources.grants regardless of whether all entries routed; any
		// unrouted entries already produced an error diagnostic above.
		newResources.SetLoc("grants", flatGrantsValue.Locations(),
			dyn.NewValue(dyn.NewMapping(), flatGrantsValue.Locations()))

		return dyn.SetByPath(root, dyn.NewPath(dyn.Key("resources")),
			dyn.NewValue(newResources, resourcesValue.Locations()))
	})
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	return diags
}
