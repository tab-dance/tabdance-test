package cmd

import (
	"fmt"
	"os"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/tab-dance/tabdance-test/internal/castconv"
)

var cast2TapeCmd = &cobra.Command{
	Use:   "cast2tape <cast> [-o out.tape]",
	Short: "Convert an asciicast recording into a deterministic tape",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cast, err := castconv.Parse(args[0])
		if err != nil {
			return fmt.Errorf("parse cast: %w", err)
		}

		tapeContent := cast.ToTapeCast()

		if output2Tape == "" {
			fmt.Print(tapeContent)
			return nil
		}
		if err := os.WriteFile(output2Tape, []byte(tapeContent), 0644); err != nil {
			return fmt.Errorf("write tape: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Wrote %s\n", output2Tape)
		return nil
	},
}

var output2Tape string

func init() {
	cast2TapeCmd.Flags().StringVarP(&output2Tape, "output", "o", "", "write the tape to this path (default: stdout)")
	rootCmd.AddCommand(cast2TapeCmd)
	carapace.Gen(cast2TapeCmd).PositionalCompletion(
		carapace.ActionFiles(".cast"),
	)
}
