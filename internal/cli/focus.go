package cli

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

func newFocusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "focus <id> <gained|lost>",
		Short: "Send focus event to a terminal",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFocus(args[0], args[1])
		},
	}

	carapace.Gen(cmd).PositionalCompletion(
		actionTerminalIDs(),
		carapace.ActionValues("gained", "lost"),
	)

	return cmd
}
