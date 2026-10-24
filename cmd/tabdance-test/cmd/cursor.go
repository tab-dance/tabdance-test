package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/tab-dance/tabdance-test/pkg/actions"
)

var cursorCmd = &cobra.Command{
	Use:   "cursor <id>",
	Short: "Print cursor position and state",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		result, err := c.Cursor(args[0])
		if err != nil {
			return err
		}

		if cursorJsonOutput {
			data, _ := json.Marshal(result)
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("x:            %d\n", result.X)
		fmt.Printf("y:            %d\n", result.Y)
		fmt.Printf("visible:      %v\n", result.Visible)
		fmt.Printf("pending_wrap: %v\n", result.PendingWrap)
		fmt.Printf("style:        %s\n", result.Style)
		return nil
	},
}

var cursorJsonOutput bool

func init() {
	cursorCmd.Flags().BoolVar(&cursorJsonOutput, "json", false, "output as JSON")
	rootCmd.AddCommand(cursorCmd)
	carapace.Gen(cursorCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
	)
}
