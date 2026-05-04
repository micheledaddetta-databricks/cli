package config

import "github.com/databricks/cli/ucm/config/resources"

// Resources is the top-level container for every UC and cloud-underlay
// resource declared in ucm.yml. For M0 only a minimal UC-native subset is
// supported; cloud resources (S3/ADLS/GCS, IAM/MI/SA, KMS) land in M2.
type Resources struct {
	Catalogs           map[string]*resources.Catalog           `json:"catalogs,omitempty"`
	Schemas            map[string]*resources.Schema            `json:"schemas,omitempty"`
	Grants             map[string]*resources.Grant             `json:"grants,omitempty"`
	StorageCredentials map[string]*resources.StorageCredential `json:"storage_credentials,omitempty"`
	ExternalLocations  map[string]*resources.ExternalLocation  `json:"external_locations,omitempty"`
	Volumes            map[string]*resources.Volume            `json:"volumes,omitempty"`
	Connections        map[string]*resources.Connection        `json:"connections,omitempty"`
	TagValidationRules map[string]*resources.TagValidationRule `json:"tag_validation_rules,omitempty"`
}

// AllGrants returns every nested grant declared anywhere under the
// resources tree, keyed by the grant's leaf key. Each returned grant
// carries Securable.{Type, Name} synthesised from its parent path when not
// already set so callers can read the securable uniformly regardless of
// whether the entry originated as flat or nested form. Synthesis is
// idempotent: pre-populated securable fields are left untouched.
//
// Consumers should prefer this helper over reading r.Grants directly:
// after FlattenNestedResources + RouteFlatGrants, r.Grants is always empty
// and the per-resource nested maps are the canonical surface.
func (r *Resources) AllGrants() map[string]*resources.Grant {
	out := map[string]*resources.Grant{}
	visit := func(kind, parentKey string, m map[string]*resources.Grant) {
		for key, g := range m {
			if g == nil {
				continue
			}
			if g.Securable.Type == "" {
				g.Securable.Type = kind
			}
			if g.Securable.Name == "" {
				g.Securable.Name = parentKey
			}
			out[key] = g
		}
	}
	for parentKey, c := range r.Catalogs {
		visit("catalog", parentKey, c.Grants)
	}
	for parentKey, s := range r.Schemas {
		visit("schema", parentKey, s.Grants)
	}
	for parentKey, v := range r.Volumes {
		visit("volume", parentKey, v.Grants)
	}
	for parentKey, e := range r.ExternalLocations {
		visit("external_location", parentKey, e.Grants)
	}
	for parentKey, sc := range r.StorageCredentials {
		visit("storage_credential", parentKey, sc.Grants)
	}
	return out
}
