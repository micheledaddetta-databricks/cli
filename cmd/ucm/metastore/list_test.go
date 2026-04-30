package metastore

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/databricks/cli/libs/flags"
	"github.com/databricks/databricks-sdk-go/service/catalog"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderMetastoresText(t *testing.T) {
	tcs := []struct {
		name       string
		metastores []catalog.MetastoreInfo
		want       string
	}{
		{
			name:       "empty",
			metastores: nil,
			want:       "No metastores found.\n",
		},
		{
			name: "multiple",
			metastores: []catalog.MetastoreInfo{
				{MetastoreId: "id-1", Name: "primary", Region: "us-west-2"},
				{MetastoreId: "id-2", Name: "secondary", Region: "us-east-1"},
			},
			want: "id-1\tprimary\tus-west-2\nid-2\tsecondary\tus-east-1\n",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := renderMetastores(&buf, tc.metastores, flags.OutputText)
			require.NoError(t, err)
			assert.Equal(t, tc.want, buf.String())
		})
	}
}

func TestRenderMetastoresJSON(t *testing.T) {
	metastores := []catalog.MetastoreInfo{
		{MetastoreId: "id-1", Name: "primary", Region: "us-west-2"},
	}

	var buf bytes.Buffer
	err := renderMetastores(&buf, metastores, flags.OutputJSON)
	require.NoError(t, err)

	// Round-trip into a generic map slice — comparing typed structs is
	// brittle because the SDK's UnmarshalJSON populates ForceSendFields
	// for every present key.
	var got []map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	require.Len(t, got, 1)
	assert.Equal(t, "id-1", got[0]["metastore_id"])
	assert.Equal(t, "primary", got[0]["name"])
	assert.Equal(t, "us-west-2", got[0]["region"])
}

func TestNewListCommandWiring(t *testing.T) {
	cmd := newListCommand()
	assert.Equal(t, "list", cmd.Use)
	assert.NotNil(t, cmd.PreRunE, "PreRunE must be set so the account client is resolved before RunE")
	assert.NotNil(t, cmd.RunE)
}

func TestNewGroupRegistersList(t *testing.T) {
	cmd := New()
	assert.Equal(t, "metastore", cmd.Use)

	var listCmd *cobra.Command
	for _, c := range cmd.Commands() {
		if c.Use == "list" {
			listCmd = c
			break
		}
	}
	require.NotNil(t, listCmd, "metastore group must register the list subcommand")
}
