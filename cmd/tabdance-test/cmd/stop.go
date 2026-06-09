package cmd

import (
	"github.com/tab-dance/tabdance-test/pkg/actions"
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop <id>",
	Short: "Stop asciinema recording",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return newClient().Stop(args[0])
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
	carapace.Gen(stopCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
	)
}
