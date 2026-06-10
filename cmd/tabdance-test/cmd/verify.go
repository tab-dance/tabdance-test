package cmd

import (
	"fmt"
	"os"

	"github.com/tab-dance/tabdance-test/pkg/actions"
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify <id> <golden>",
	Short: "Verify terminal output against a golden file",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		result, err := c.Verify(args[0], args[1], verifyFormat, verifyTrim, verifyUnwrap)
		if err != nil {
			return err
		}

		if result.Match {
			fmt.Println("OK: output matches golden file")
			return nil
		}

		fmt.Fprintf(os.Stderr, "FAIL: output does not match golden file\n")
		fmt.Fprintf(os.Stderr, "%s\n", result.Diff)
		os.Exit(1)
		return nil
	},
}

var (
	verifyFormat string
	verifyTrim   bool
	verifyUnwrap bool
)

func init() {
	verifyCmd.Flags().StringVar(&verifyFormat, "format", "plain", "output format: plain, vt, html, svg, json")
	verifyCmd.Flags().BoolVar(&verifyTrim, "trim", true, "trim trailing whitespace")
	verifyCmd.Flags().BoolVar(&verifyUnwrap, "unwrap", true, "unwrap soft-wrapped lines")
	rootCmd.AddCommand(verifyCmd)
	carapace.Gen(verifyCmd).FlagCompletion(carapace.ActionMap{
		"format": carapace.ActionValues("plain", "vt", "html", "svg", "json"),
	})
	carapace.Gen(verifyCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
		carapace.ActionFiles(),
	)
}
