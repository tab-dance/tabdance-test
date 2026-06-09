package cmd

import (
	"github.com/tab-dance/tabdance-test/pkg/actions"
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

var pasteCmd = &cobra.Command{
	Use:   "paste <id> <text>",
	Short: "Send paste event (bracketed-paste-aware) to a terminal",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return newClient().Paste(args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(pasteCmd)
	carapace.Gen(pasteCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
	)
}
