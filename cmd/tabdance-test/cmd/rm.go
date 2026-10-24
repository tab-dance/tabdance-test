package cmd

import (
	"fmt"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/tab-dance/tabdance-test/pkg/actions"
)

var rmCmd = &cobra.Command{
	Use:   "rm <id>",
	Short: "Remove and close a terminal",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := newClient().Rm(args[0]); err != nil {
			return err
		}
		fmt.Printf("Terminal %s removed\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
	carapace.Gen(rmCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
	)
}
