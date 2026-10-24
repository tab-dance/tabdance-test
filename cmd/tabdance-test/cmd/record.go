package cmd

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/tab-dance/tabdance-test/pkg/actions"
)

var recordCmd = &cobra.Command{
	Use:   "record <id> <path>",
	Short: "Start asciinema recording for a terminal",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return newClient().Record(args[0], args[1], recordIdleTimeLimit)
	},
}

var recordIdleTimeLimit float64

func init() {
	recordCmd.Flags().Float64Var(&recordIdleTimeLimit, "idle-time-limit", 2.0, "idle time limit in seconds")
	rootCmd.AddCommand(recordCmd)
	carapace.Gen(recordCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
		carapace.ActionFiles(".cast"),
	)
}
