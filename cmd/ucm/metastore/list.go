package metastore

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/databricks/cli/cmd/root"
	"github.com/databricks/cli/cmd/ucm/utils"
	"github.com/databricks/cli/libs/cmdctx"
	"github.com/databricks/cli/libs/flags"
	"github.com/databricks/databricks-sdk-go/service/catalog"
	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List metastores accessible to the configured account.",
		Long:    `List metastores accessible to the configured account. Requires account-admin scope.`,
		Args:    root.NoArgs,
		PreRunE: utils.MustAccountClient,
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		a := cmdctx.AccountClient(ctx)

		metastores, err := a.Metastores.ListAll(ctx)
		if err != nil {
			return fmt.Errorf("list metastores: %w", err)
		}

		return renderMetastores(cmd.OutOrStdout(), metastores, root.OutputType(cmd))
	}

	return cmd
}

// renderMetastores emits the metastore listing as text or JSON. Extracted so
// tests can exercise it without wiring the account-client PreRunE chain.
func renderMetastores(out io.Writer, metastores []catalog.MetastoreInfo, output flags.Output) error {
	if output == flags.OutputJSON {
		return renderJSON(out, metastores)
	}
	return renderText(out, metastores)
}

func renderText(out io.Writer, metastores []catalog.MetastoreInfo) error {
	if len(metastores) == 0 {
		_, err := fmt.Fprintln(out, "No metastores found.")
		return err
	}
	for _, m := range metastores {
		if _, err := fmt.Fprintf(out, "%s\t%s\t%s\n", m.MetastoreId, m.Name, m.Region); err != nil {
			return err
		}
	}
	return nil
}

func renderJSON(out io.Writer, metastores []catalog.MetastoreInfo) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(metastores)
}
