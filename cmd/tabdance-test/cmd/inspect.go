package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/tab-dance/tabdance-test/pkg/actions"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect <id>",
	Short: "Inspect terminal details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		result, err := c.Inspect(args[0])
		if err != nil {
			return err
		}

		if inspectJsonOutput {
			data, _ := json.Marshal(result)
			fmt.Println(string(data))
			return nil
		}

		for k, v := range result {
			fmt.Printf("%s: %v\n", k, v)
		}
		return nil
	},
}

var inspectJsonOutput bool

func init() {
	inspectCmd.Flags().BoolVar(&inspectJsonOutput, "json", false, "output as JSON")
	rootCmd.AddCommand(inspectCmd)
	carapace.Gen(inspectCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
	)
}
