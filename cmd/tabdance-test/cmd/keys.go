package cmd

import (
	"github.com/tab-dance/tabdance-test/pkg/actions"
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

var keysCmd = &cobra.Command{
	Use:   "keys <id> <key...>",
	Short: "Send key events to a terminal",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return newClient().Keys(args[0], args[1:])
	},
}

func init() {
	rootCmd.AddCommand(keysCmd)
	carapace.Gen(keysCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
		actions.ActionKeyNames(),
	)
	carapace.Gen(keysCmd).PositionalAnyCompletion(
		actions.ActionKeyNames(),
	)
}
