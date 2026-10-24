package cmd

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/tab-dance/tabdance-test/pkg/actions"
)

var tapeCmd = &cobra.Command{
	Use:   "tape <id> <path>",
	Short: "Play a VHS tape file in a terminal",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return newClient().Tape(args[0], args[1], tapeTimeout)
	},
}

var tapeTimeout float64

func init() {
	tapeCmd.Flags().Float64Var(&tapeTimeout, "timeout", 30, "timeout in seconds for Wait commands")
	rootCmd.AddCommand(tapeCmd)
	carapace.Gen(tapeCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
		carapace.ActionFiles(".tape"),
	)
}
