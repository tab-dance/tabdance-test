package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/tab-dance/tabdance-test/pkg/actions"
)

var verifyCmd = &cobra.Command{
	Use:   "verify <id> <golden>",
	Short: "Verify terminal output against a golden file",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		result, err := c.Verify(args[0], args[1], outFormat, outTrim, outUnwrap)
		if err != nil {
			return err
		}

		if verifyJsonOutput {
			data, _ := json.Marshal(result)
			fmt.Println(string(data))
			if !result.Match {
				os.Exit(1)
			}
			return nil
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

var verifyJsonOutput bool

func init() {
	verifyCmd.Flags().BoolVar(&verifyJsonOutput, "json", false, "output as JSON")
	addOutputFlags(verifyCmd, allFormats)
	rootCmd.AddCommand(verifyCmd)
	carapace.Gen(verifyCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
		carapace.ActionFiles(),
	)
}
