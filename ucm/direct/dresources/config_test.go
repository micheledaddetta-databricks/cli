package dresources

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustLoadConfig(t *testing.T) {
	cfg := MustLoadConfig()
	assert.NotNil(t, cfg)
}

func TestGetResourceConfig(t *testing.T) {
	// Nonexistent resources return an empty config without panicking.
	assert.Empty(t, GetResourceConfig("nonexistent").RecreateOnChanges)
}

// TestVolumeRecreateOnImmutableFields covers #62: UpdateVolumeRequestContent
// only accepts {comment, new_name, owner}, so any change to a volume's
// catalog_name, schema_name, storage_location, or volume_type would be
// silently dropped on Update — without recreate_on_changes the planner loops.
func TestVolumeRecreateOnImmutableFields(t *testing.T) {
	cfg := GetResourceConfig("volumes")
	got := make([]string, 0, len(cfg.RecreateOnChanges))
	for _, r := range cfg.RecreateOnChanges {
		got = append(got, r.Field.String())
	}
	assert.ElementsMatch(t, []string{
		"catalog_name",
		"schema_name",
		"storage_location",
		"volume_type",
	}, got)
}

// TestConnectionRecreateOnImmutableFields covers #62 for connections:
// UpdateConnection accepts only {new_name, options, owner}, so changes to
// connection_type / comment / properties / read_only must trigger recreate.
func TestConnectionRecreateOnImmutableFields(t *testing.T) {
	cfg := GetResourceConfig("connections")
	got := make([]string, 0, len(cfg.RecreateOnChanges))
	for _, r := range cfg.RecreateOnChanges {
		got = append(got, r.Field.String())
	}
	assert.ElementsMatch(t, []string{
		"connection_type",
		"comment",
		"properties",
		"read_only",
	}, got)
}
