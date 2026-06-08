package cli

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

func newPasteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "paste <id> <text>",
		Short: "Send paste event (bracketed-paste-aware) to a terminal",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPaste(args[0], args[1])
		},
	}

	carapace.Gen(cmd).PositionalCompletion(
		actionTerminalIDs(),
	)

	return cmd
}
