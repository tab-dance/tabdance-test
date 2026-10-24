package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/tab-dance/tabdance-test/pkg/actions"
)

var viewCmd = &cobra.Command{
	Use:   "view <id>",
	Short: "Print terminal screen content",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		result, err := c.View(args[0], outFormat, outTrim, outUnwrap)
		if err != nil {
			return err
		}

		if viewJsonOutput {
			data, _ := json.Marshal(result)
			fmt.Println(string(data))
			return nil
		}

		fmt.Print(result.Content)
		return nil
	},
}

var viewJsonOutput bool

func init() {
	viewCmd.Flags().BoolVar(&viewJsonOutput, "json", false, "output as JSON")
	addOutputFlags(viewCmd, allFormats)
	rootCmd.AddCommand(viewCmd)
	carapace.Gen(viewCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
	)
}
