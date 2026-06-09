package cmd

import (
	"github.com/tab-dance/tabdance-test/pkg/actions"
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

var attachCmd = &cobra.Command{
	Use:   "attach <id>",
	Short: "Attach to terminal's PTY output (streaming)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return newClient().Attach(args[0])
	},
}

func init() {
	rootCmd.AddCommand(attachCmd)
	carapace.Gen(attachCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
	)
}
