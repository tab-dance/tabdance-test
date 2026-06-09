package cmd

import (
	"github.com/tab-dance/tabdance-test/pkg/actions"
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

var keysMods string

var keysCmd = &cobra.Command{
	Use:   "keys <id> <key...>",
	Short: "Send key events to a terminal",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return newClient().Keys(args[0], args[1:], keysMods)
	},
}

func init() {
	keysCmd.Flags().StringVar(&keysMods, "mods", "", "modifiers to apply to all keys (e.g. \"ctrl\", \"ctrl+alt\")")
	rootCmd.AddCommand(keysCmd)
	carapace.Gen(keysCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
		actions.ActionKeyNames(),
	)
	carapace.Gen(keysCmd).PositionalAnyCompletion(
		actions.ActionKeyNames(),
	)
	carapace.Gen(keysCmd).FlagCompletion(carapace.ActionMap{
		"mods": carapace.ActionValues("ctrl", "alt", "shift", "super", "ctrl+alt", "ctrl+shift", "alt+shift", "ctrl+alt+shift"),
	})
}
