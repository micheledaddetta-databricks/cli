// Package metastore wires the `databricks ucm metastore` subcommand group:
// account-scoped read/write operations on Unity Catalog metastores. Mirrors
// cmd/ucm/deployment/ in shape but forks rather than imports cmd/account/
// metastores (auto-generated upstream code).
package metastore

import (
	"github.com/spf13/cobra"
)

// New returns the cobra group registered under `databricks ucm metastore`.
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metastore",
		Short: "Account-scoped UC metastore operations",
		Long: `Account-scoped Unity Catalog metastore operations.

These commands authenticate against the Databricks account API (not the
workspace API) and require the configured service principal to hold
account-admin scope. Configure the account host via ucm.yml's
workspace.account_host field, or via the DATABRICKS_HOST env var with an
account-style URL.`,
	}

	cmd.AddCommand(newListCommand())
	return cmd
}
