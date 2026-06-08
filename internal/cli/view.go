package cli

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

func newViewCmd() *cobra.Command {
	var format string
	var trim bool
	var unwrap bool
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "Print terminal screen content",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runView(args[0], format, trim, unwrap, jsonOutput)
		},
	}

	cmd.Flags().StringVar(&format, "format", "plain", "output format: plain, vt, html")
	cmd.Flags().BoolVar(&trim, "trim", true, "trim trailing whitespace")
	cmd.Flags().BoolVar(&unwrap, "unwrap", true, "unwrap soft-wrapped lines")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")

	carapace.Gen(cmd).FlagCompletion(carapace.ActionMap{
		"format": carapace.ActionValues("plain", "vt", "html"),
	})

	carapace.Gen(cmd).PositionalCompletion(
		actionTerminalIDs(),
	)

	return cmd
}
