package cli

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

func newCellCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "cell <id> <x> <y>",
		Short: "Inspect a specific cell's content and style",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCell(args[0], args[1], args[2], jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")

	carapace.Gen(cmd).PositionalCompletion(
		actionTerminalIDs(),
	)

	return cmd
}
