package cmd

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/tab-dance/tabdance-test/pkg/actions"
)

var markerCmd = &cobra.Command{
	Use:   "marker <id> <label>",
	Short: "Add a marker to the current recording",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return newClient().Marker(args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(markerCmd)
	carapace.Gen(markerCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
	)
}
