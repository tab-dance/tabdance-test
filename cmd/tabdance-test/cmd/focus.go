package cmd

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/tab-dance/tabdance-test/pkg/actions"
)

var focusCmd = &cobra.Command{
	Use:   "focus <id> <gained|lost>",
	Short: "Send focus event to a terminal",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return newClient().Focus(args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(focusCmd)
	carapace.Gen(focusCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
		carapace.ActionValues("gained", "lost"),
	)
}
