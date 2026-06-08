package cli

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

func newMouseCmd() *cobra.Command {
	var mods string

	cmd := &cobra.Command{
		Use:   "mouse <id> <action> <button> <x> <y>",
		Short: "Send mouse events to a terminal",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMouse(args[0], args[1], args[2], args[3], args[4], mods)
		},
	}

	cmd.Flags().StringVar(&mods, "mods", "", "modifiers: ctrl, alt, shift, super (combinable with +)")

	carapace.Gen(cmd).PositionalCompletion(
		actionTerminalIDs(),
		carapace.ActionValues("press", "release", "move", "scroll_up", "scroll_down"),
		carapace.ActionValues("left", "middle", "right"),
	)

	return cmd
}
