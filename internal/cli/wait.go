package cli

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

func newWaitCmd() *cobra.Command {
	var scope string
	var timeout float64
	var format string
	var trim bool
	var unwrap bool

	cmd := &cobra.Command{
		Use:   "wait <id> <pattern>",
		Short: "Wait for a regex pattern on the terminal screen",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWait(args[0], args[1], scope, timeout, format, trim, unwrap)
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "screen", "match scope: screen or line")
	cmd.Flags().Float64Var(&timeout, "timeout", 10, "timeout in seconds")
	cmd.Flags().StringVar(&format, "format", "plain", "output format: plain, vt, html")
	cmd.Flags().BoolVar(&trim, "trim", true, "trim trailing whitespace")
	cmd.Flags().BoolVar(&unwrap, "unwrap", true, "unwrap soft-wrapped lines")

	carapace.Gen(cmd).FlagCompletion(carapace.ActionMap{
		"scope":  carapace.ActionValues("screen", "line"),
		"format": carapace.ActionValues("plain", "vt", "html"),
	})

	carapace.Gen(cmd).PositionalCompletion(
		actionTerminalIDs(),
	)

	return cmd
}
