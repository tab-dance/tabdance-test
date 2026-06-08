package cli

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

func newVerifyCmd() *cobra.Command {
	var format string
	var trim bool
	var unwrap bool

	cmd := &cobra.Command{
		Use:   "verify <id> <golden>",
		Short: "Verify terminal output against a golden file",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVerify(args[0], args[1], format, trim, unwrap)
		},
	}

	cmd.Flags().StringVar(&format, "format", "plain", "output format: plain, vt, html")
	cmd.Flags().BoolVar(&trim, "trim", true, "trim trailing whitespace")
	cmd.Flags().BoolVar(&unwrap, "unwrap", true, "unwrap soft-wrapped lines")

	carapace.Gen(cmd).FlagCompletion(carapace.ActionMap{
		"format": carapace.ActionValues("plain", "vt", "html"),
	})

	carapace.Gen(cmd).PositionalCompletion(
		actionTerminalIDs(),
		carapace.ActionFiles(),
	)

	return cmd
}
