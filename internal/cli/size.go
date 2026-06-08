package cli

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

func newSizeCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "size <id>",
		Short: "Print terminal dimensions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSize(args[0], jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")

	carapace.Gen(cmd).PositionalCompletion(
		actionTerminalIDs(),
	)

	return cmd
}
