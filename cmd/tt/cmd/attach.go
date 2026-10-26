package cmd

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/tab-dance/tabdance-test/pkg/actions"
)

var attachCmd = &cobra.Command{
	Use:   "attach <id>",
	Short: "Stream raw PTY output from attach time (Ctrl+C to detach)",
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
