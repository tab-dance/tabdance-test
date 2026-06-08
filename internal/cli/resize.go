package cli

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

func newResizeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resize <id> <cols> <rows>",
		Short: "Resize a terminal",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runResize(args[0], args[1], args[2])
		},
	}

	carapace.Gen(cmd).PositionalCompletion(
		actionTerminalIDs(),
	)

	return cmd
}
