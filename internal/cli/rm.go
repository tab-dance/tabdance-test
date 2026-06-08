package cli

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

func newRmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm <id>",
		Short: "Remove and close a terminal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRm(args[0])
		},
	}

	carapace.Gen(cmd).PositionalCompletion(
		actionTerminalIDs(),
	)

	return cmd
}
