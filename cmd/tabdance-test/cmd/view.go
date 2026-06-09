package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tab-dance/tabdance-test/pkg/actions"
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view <id>",
	Short: "Print terminal screen content",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		result, err := c.View(args[0], viewFormat, viewTrim, viewUnwrap, viewOutputPath)
		if err != nil {
			return err
		}

		if viewJsonOutput {
			data, _ := json.Marshal(result)
			fmt.Println(string(data))
			return nil
		}

		if result.OutputPath != "" {
			data, err := os.ReadFile(result.OutputPath)
			if err != nil {
				return fmt.Errorf("read output file: %w", err)
			}
			os.Stdout.Write(data)
			return nil
		}

		fmt.Print(result.Content)
		return nil
	},
}

var (
	viewFormat     string
	viewTrim       bool
	viewUnwrap     bool
	viewJsonOutput bool
	viewOutputPath string
)

func init() {
	viewCmd.Flags().StringVar(&viewFormat, "format", "plain", "output format: plain, vt, html, svg, png")
	viewCmd.Flags().BoolVar(&viewTrim, "trim", true, "trim trailing whitespace")
	viewCmd.Flags().BoolVar(&viewUnwrap, "unwrap", true, "unwrap soft-wrapped lines")
	viewCmd.Flags().BoolVar(&viewJsonOutput, "json", false, "output as JSON")
	viewCmd.Flags().StringVar(&viewOutputPath, "output", "", "output file path for image formats (svg, png)")
	rootCmd.AddCommand(viewCmd)
	carapace.Gen(viewCmd).FlagCompletion(carapace.ActionMap{
		"format": carapace.ActionValues("plain", "vt", "html", "svg", "png"),
	})
	carapace.Gen(viewCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
	)
}
