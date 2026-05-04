package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultBuildPhaseOptions_NilUcmReturnsCleanError covers the defensive
// guard added for #27. Production callers gate on a non-nil *ucm.Ucm before
// invoking BuildPhaseOptionsHook, but library-mode embeddings or test
// shortcuts that skip the harness can reach this path with u == nil.
// Without the guard the function would panic on `u.WorkspaceClientE()`.
func TestDefaultBuildPhaseOptions_NilUcmReturnsCleanError(t *testing.T) {
	_, err := DefaultBuildPhaseOptions(t.Context(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ucm:")
	assert.Contains(t, err.Error(), "ProcessUcm")
}
