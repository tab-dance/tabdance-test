package cmd

import (
	"fmt"
	"os"

	"github.com/tab-dance/tabdance-test/pkg/actions"
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot <id> <golden>",
	Short: "Update or create a golden file with current terminal output",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := newClient().Snapshot(args[0], args[1], snapshotFormat, snapshotTrim, snapshotUnwrap); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Golden file updated: %s\n", args[1])
		return nil
	},
}

var (
	snapshotFormat string
	snapshotTrim   bool
	snapshotUnwrap bool
)

func init() {
	snapshotCmd.Flags().StringVar(&snapshotFormat, "format", "plain", "output format: plain, vt, html")
	snapshotCmd.Flags().BoolVar(&snapshotTrim, "trim", true, "trim trailing whitespace")
	snapshotCmd.Flags().BoolVar(&snapshotUnwrap, "unwrap", true, "unwrap soft-wrapped lines")
	rootCmd.AddCommand(snapshotCmd)
	carapace.Gen(snapshotCmd).FlagCompletion(carapace.ActionMap{
		"format": carapace.ActionValues("plain", "vt", "html"),
	})
	carapace.Gen(snapshotCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
		carapace.ActionFiles(),
	)
}
