package cmd

import (
	"fmt"
	"strconv"

	"github.com/tab-dance/tabdance-test/pkg/actions"
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

var mouseCmd = &cobra.Command{
	Use:   "mouse <id> <action> <button> <x> <y>",
	Short: "Send mouse events to a terminal",
	Args:  cobra.ExactArgs(5),
	RunE: func(cmd *cobra.Command, args []string) error {
		x, err := strconv.ParseFloat(args[3], 32)
		if err != nil {
			return fmt.Errorf("invalid x: %w", err)
		}
		y, err := strconv.ParseFloat(args[4], 32)
		if err != nil {
			return fmt.Errorf("invalid y: %w", err)
		}

		return newClient().Mouse(args[0], args[1], args[2], float32(x), float32(y), mouseMods)
	},
}

var mouseMods string

func init() {
	mouseCmd.Flags().StringVar(&mouseMods, "mods", "", "modifiers: ctrl, alt, shift, super (combinable with +)")
	rootCmd.AddCommand(mouseCmd)
	carapace.Gen(mouseCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
		carapace.ActionValues("press", "release", "move", "scroll_up", "scroll_down"),
		carapace.ActionValues("left", "middle", "right"),
	)
}
