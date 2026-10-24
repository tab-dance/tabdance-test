package cmd

import (
	"fmt"
	"strconv"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/tab-dance/tabdance-test/pkg/actions"
)

var resizeCmd = &cobra.Command{
	Use:   "resize <id> <cols> <rows>",
	Short: "Resize a terminal",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		cols, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid cols: %w", err)
		}
		rows, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("invalid rows: %w", err)
		}

		return newClient().Resize(args[0], cols, rows)
	},
}

func init() {
	rootCmd.AddCommand(resizeCmd)
	carapace.Gen(resizeCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
	)
}
