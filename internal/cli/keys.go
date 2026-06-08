package cli

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

func newKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys <id> <key...>",
		Short: "Send key events to a terminal",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runKeys(args[0], args[1:])
		},
	}

	carapace.Gen(cmd).PositionalCompletion(
		actionTerminalIDs(),
		actionKeyNames(),
	)

	carapace.Gen(cmd).PositionalAnyCompletion(
		actionKeyNames(),
	)

	return cmd
}
