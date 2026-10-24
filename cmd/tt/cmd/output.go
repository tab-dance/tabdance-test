package cmd

import (
	"fmt"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/tab-dance/tabdance-test/pkg/actions"
)

var outputCmd = &cobra.Command{
	Use:   "output <id> [bytes]",
	Short: "Print the terminal's recent raw PTY output (control chars visible)",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		n := 0
		if len(args) == 2 {
			if _, err := fmt.Sscanf(args[1], "%d", &n); err != nil {
				return fmt.Errorf("bytes %q is not a number", args[1])
			}
		}
		out, err := newClient().Output(args[0], n)
		if err != nil {
			return err
		}
		fmt.Print(out)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(outputCmd)
	carapace.Gen(outputCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
	)
}
