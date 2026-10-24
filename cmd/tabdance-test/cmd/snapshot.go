package cmd

import (
	"fmt"
	"os"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/tab-dance/tabdance-test/pkg/actions"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot <id> <golden>",
	Short: "Update or create a golden file with current terminal output",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := newClient().Snapshot(args[0], args[1], outFormat, outTrim, outUnwrap); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Golden file updated: %s\n", args[1])
		return nil
	},
}

func init() {
	addOutputFlags(snapshotCmd, allFormats)
	rootCmd.AddCommand(snapshotCmd)
	carapace.Gen(snapshotCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
		carapace.ActionFiles(),
	)
}
