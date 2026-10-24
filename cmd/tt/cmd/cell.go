package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/tab-dance/tabdance-test/pkg/actions"
)

var cellCmd = &cobra.Command{
	Use:   "cell <id> <x> <y>",
	Short: "Inspect a specific cell's content and style",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		x, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid x: %w", err)
		}
		y, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("invalid y: %w", err)
		}

		result, err := c.Cell(args[0], x, y)
		if err != nil {
			return err
		}

		if cellJsonOutput {
			data, _ := json.Marshal(result)
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("codepoint: %d\n", result.Codepoint)
		fmt.Printf("text:      %s\n", result.Text)
		fmt.Printf("wide:      %v\n", result.Wide)
		fmt.Printf("bold:      %v\n", result.Bold)
		fmt.Printf("italic:    %v\n", result.Italic)
		fmt.Printf("fg:        %s\n", result.FgColor)
		fmt.Printf("bg:        %s\n", result.BgColor)
		fmt.Printf("inverse:   %v\n", result.Inverse)
		if result.Hyperlink != "" {
			fmt.Printf("hyperlink: %s\n", result.Hyperlink)
		}
		return nil
	},
}

var cellJsonOutput bool

func init() {
	cellCmd.Flags().BoolVar(&cellJsonOutput, "json", false, "output as JSON")
	rootCmd.AddCommand(cellCmd)
	carapace.Gen(cellCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
	)
}
