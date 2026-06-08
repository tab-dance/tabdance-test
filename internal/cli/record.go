package cli

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

func newRecordCmd() *cobra.Command {
	var idleTimeLimit float64

	cmd := &cobra.Command{
		Use:   "record <id> <path>",
		Short: "Start asciinema recording for a terminal",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRecord(args[0], args[1], idleTimeLimit)
		},
	}

	cmd.Flags().Float64Var(&idleTimeLimit, "idle-time-limit", 2.0, "idle time limit in seconds")

	carapace.Gen(cmd).PositionalCompletion(
		actionTerminalIDs(),
		carapace.ActionFiles(".cast"),
	)

	return cmd
}
