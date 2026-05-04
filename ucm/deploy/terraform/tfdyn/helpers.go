package tfdyn

import (
	"sort"

	"github.com/databricks/cli/libs/dyn"
)

// appendString adds key=fallback to pairs when vin does not already define
// the field. If vin sets the field, the caller's value wins and its
// location is preserved.
func appendString(pairs *[]dyn.Pair, vin dyn.Value, key, fallback string) {
	v := vin.Get(key)
	if s, ok := v.AsString(); ok && s != "" {
		*pairs = append(*pairs, dyn.Pair{
			Key:   dyn.NewValue(key, v.Locations()),
			Value: v,
		})
		return
	}
	*pairs = append(*pairs, dyn.Pair{
		Key:   dyn.V(key),
		Value: dyn.V(fallback),
	})
}

// appendStringIfSet appends key=vin[key] when the field is a non-empty
// string. Missing or empty fields are skipped so the Terraform JSON does
// not carry empty `"comment": ""` noise.
func appendStringIfSet(pairs *[]dyn.Pair, vin dyn.Value, key string) {
	v := vin.Get(key)
	s, ok := v.AsString()
	if !ok || s == "" {
		return
	}
	*pairs = append(*pairs, dyn.Pair{
		Key:   dyn.NewValue(key, v.Locations()),
		Value: v,
	})
}

// appendBoolIfSet emits key=vin[key] when vin[key] is a bool true. A false
// value (the zero) is skipped so the Terraform JSON stays clean.
func appendBoolIfSet(pairs *[]dyn.Pair, vin dyn.Value, key string) {
	v := vin.Get(key)
	b, ok := v.AsBool()
	if !ok || !b {
		return
	}
	*pairs = append(*pairs, dyn.Pair{
		Key:   dyn.NewValue(key, v.Locations()),
		Value: v,
	})
}

// mapFromValue returns v as a map-typed dyn.Value with pairs sorted by key
// when v holds at least one entry; otherwise the second return is false so
// callers can skip emitting an empty map. Sorting matters because typed
// config Tags / Properties / Options are Go maps (`map[string]string`)
// whose iteration order leaks into convert.Normalize. Without it the
// generated tf.json drifts byte-for-byte across runs.
func mapFromValue(v dyn.Value) (dyn.Value, bool) {
	m, ok := v.AsMap()
	if !ok || m.Len() == 0 {
		return dyn.InvalidValue, false
	}
	return sortMapByKeys(v), true
}

// sortMapByKeys returns a copy of v whose pairs are sorted lexicographically
// by key. Non-map values are returned unchanged. The sort is stable and
// deterministic so callers can build byte-identical tf.json across runs.
func sortMapByKeys(v dyn.Value) dyn.Value {
	m, ok := v.AsMap()
	if !ok {
		return v
	}
	pairs := m.Pairs()
	sorted := make([]dyn.Pair, len(pairs))
	copy(sorted, pairs)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Key.MustString() < sorted[j].Key.MustString()
	})
	return dyn.NewValue(dyn.NewMappingFromPairs(sorted), v.Locations())
}
