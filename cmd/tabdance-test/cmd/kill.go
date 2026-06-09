package cmd

import (
	"fmt"

	"github.com/tab-dance/tabdance-test/pkg/actions"
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

var killCmd = &cobra.Command{
	Use:   "kill <id>",
	Short: "Kill a terminal",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := newClient().Kill(args[0]); err != nil {
			return err
		}
		fmt.Printf("Terminal %s killed\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(killCmd)
	carapace.Gen(killCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
	)
}
