package cli

import (
	"github.com/spf13/cobra"
)

func newLsCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List active terminals",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLs(jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")

	return cmd
}
