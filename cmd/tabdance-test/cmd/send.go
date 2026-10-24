package cmd

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/tab-dance/tabdance-test/pkg/actions"
)

var sendCmd = &cobra.Command{
	Use:   "send <id> <text>",
	Short: "Send raw text to a terminal's PTY stdin",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return newClient().Send(args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(sendCmd)
	carapace.Gen(sendCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
	)
}
