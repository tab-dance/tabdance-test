package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/tab-dance/tabdance-test/pkg/actions"
)

var waitCmd = &cobra.Command{
	Use:   "wait <id> <pattern>",
	Short: "Wait for a regex pattern on the terminal screen",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		result, err := c.Wait(args[0], args[1], waitScope, waitTimeout, outFormat, outTrim, outUnwrap)
		if err != nil {
			return err
		}

		if waitJsonOutput {
			data, _ := json.Marshal(result)
			fmt.Println(string(data))
			if !result.Found {
				os.Exit(1)
			}
			return nil
		}

		if result.Found {
			fmt.Printf("Pattern found in %dms\n", result.ElapsedMs)
			return nil
		}

		fmt.Fprintf(os.Stderr, "Pattern not found after %.0fs\n", waitTimeout)
		os.Exit(1)
		return nil
	},
}

var (
	waitScope      string
	waitTimeout    float64
	waitJsonOutput bool
)

func init() {
	waitCmd.Flags().StringVar(&waitScope, "scope", "screen", "match scope: screen or line")
	waitCmd.Flags().Float64Var(&waitTimeout, "timeout", 10, "timeout in seconds")
	waitCmd.Flags().BoolVar(&waitJsonOutput, "json", false, "output as JSON")
	addOutputFlags(waitCmd, textFormats)
	rootCmd.AddCommand(waitCmd)
	carapace.Gen(waitCmd).FlagCompletion(carapace.ActionMap{
		"scope": carapace.ActionValues("screen", "line"),
	})
	carapace.Gen(waitCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
	)
}
