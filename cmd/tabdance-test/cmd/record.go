package cmd

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/tab-dance/tabdance-test/pkg/actions"
)

var recordCmd = &cobra.Command{
	Use:   "record <id> <path>",
	Short: "Start asciinema recording for a terminal",
	Args: func(cmd *cobra.Command, args []string) error {
		if recordStop {
			return cobra.ExactArgs(1)(cmd, args)
		}
		return cobra.ExactArgs(2)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if recordStop {
			return newClient().Stop(args[0])
		}
		return newClient().Record(args[0], args[1], recordIdleTimeLimit)
	},
}

var (
	recordIdleTimeLimit float64
	recordStop          bool
)

func init() {
	recordCmd.Flags().Float64Var(&recordIdleTimeLimit, "idle-time-limit", 2.0, "idle time limit in seconds")
	recordCmd.Flags().BoolVar(&recordStop, "stop", false, "stop the terminal's recording instead of starting one")
	rootCmd.AddCommand(recordCmd)
	carapace.Gen(recordCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
		carapace.ActionFiles(".cast"),
	)
}
