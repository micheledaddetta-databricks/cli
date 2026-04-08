package experimental

import (
	aitoolscmd "github.com/databricks/cli/experimental/aitools/cmd"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "experimental",
		Short:  "Experimental commands that may change in future versions",
		Hidden: true,
		Long: `Experimental commands that may change in future versions.

╔════════════════════════════════════════════════════════════════╗
║  ⚠️  EXPERIMENTAL: These commands may change in future versions ║
╚════════════════════════════════════════════════════════════════╝

These commands provide early access to new features that are still under
development. They may change or be removed in future versions without notice.`,
	}

	// Keep aitools under experimental as a hidden backward-compatibility alias.
	// The primary command is now registered at the top level.
	aitoolsAlias := aitoolscmd.NewAitoolsCmd()
	aitoolsAlias.Hidden = true
	aitoolsAlias.Deprecated = "use 'databricks aitools' instead"
	cmd.AddCommand(aitoolsAlias)

	return cmd
}
