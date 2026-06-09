package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/tab-dance/tabdance-test/pkg/actions"
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

var sizeCmd = &cobra.Command{
	Use:   "size <id>",
	Short: "Print terminal dimensions",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		result, err := c.Size(args[0])
		if err != nil {
			return err
		}

		if sizeJsonOutput {
			data, _ := json.Marshal(result)
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("cols:      %d\n", result.Cols)
		fmt.Printf("rows:      %d\n", result.Rows)
		fmt.Printf("width_px:  %d\n", result.WidthPx)
		fmt.Printf("height_px: %d\n", result.HeightPx)
		return nil
	},
}

var sizeJsonOutput bool

func init() {
	sizeCmd.Flags().BoolVar(&sizeJsonOutput, "json", false, "output as JSON")
	rootCmd.AddCommand(sizeCmd)
	carapace.Gen(sizeCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
	)
}
