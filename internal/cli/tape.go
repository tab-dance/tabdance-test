package cli

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

func newTapeCmd() *cobra.Command {
	var timeout float64

	cmd := &cobra.Command{
		Use:   "tape <id> <path>",
		Short: "Play a VHS tape file in a terminal",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTape(args[0], args[1], timeout)
		},
	}

	cmd.Flags().Float64Var(&timeout, "timeout", 30, "timeout in seconds for Wait commands")

	carapace.Gen(cmd).PositionalCompletion(
		actionTerminalIDs(),
		carapace.ActionFiles(".tape"),
	)

	return cmd
}
