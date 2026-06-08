package cli

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

func newMarkerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "marker <id> <label>",
		Short: "Add a marker to the current recording",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMarker(args[0], args[1])
		},
	}

	carapace.Gen(cmd).PositionalCompletion(
		actionTerminalIDs(),
	)

	return cmd
}
