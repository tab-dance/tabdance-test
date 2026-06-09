package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/tab-dance/tabdance-test/internal/golden"
	"github.com/tab-dance/tabdance-test/internal/tape"
	"github.com/tab-dance/tabdance-test/internal/terminal"
)

var runCmd = &cobra.Command{
	Use:   "run <tape>",
	Short: "Play a tape file, verify against golden, exit 0/1 (CI-friendly)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		env := make(map[string]string)
		for _, e := range runEnv {
			k, v, _ := parseKeyValue(e)
			env[k] = v
		}

		mt, err := terminal.NewManagedTerminal("run_1", terminal.CreateConfig{
			Cols:             uint16(runCols),
			Rows:             uint16(runRows),
			Scrollback:       10000,
			Command:          "sh",
			Env:              env,
			WorkingDirectory: runWorkingDir,
		})
		if err != nil {
			return fmt.Errorf("create terminal: %w", err)
		}
		defer mt.Close()

		if runRecordPath != "" {
			if err := mt.StartRecording(runRecordPath, 2.0); err != nil {
				return fmt.Errorf("start recording: %w", err)
			}
		}

		executor := tape.NewTapeExecutor(mt, time.Duration(runTimeout*float64(time.Second)))
		if err := executor.Execute(args[0]); err != nil {
			return fmt.Errorf("execute tape: %w", err)
		}

		if mt.Recorder != nil {
			mt.StopRecording()
		}

		if runGoldenPath != "" {
			if runUpdate {
				mt.Lock()
				err := golden.UpdateGolden(mt.Terminal, runGoldenPath, "plain", true, true)
				mt.Unlock()
				if err != nil {
					return err
				}
				fmt.Fprintf(os.Stderr, "UPDATED %s\n", runGoldenPath)
				return nil
			}

			mt.Lock()
			result, err := golden.VerifyOutput(mt.Terminal, runGoldenPath, "plain", true, true)
			mt.Unlock()
			if err != nil {
				return err
			}

			if result.Match {
				fmt.Fprintf(os.Stderr, "PASS\n")
				return nil
			}

			fmt.Fprintf(os.Stderr, "FAIL\n%s\n", result.Diff)
			os.Exit(1)
		}

		return nil
	},
}

var (
	runGoldenPath string
	runUpdate     bool
	runCols       int
	runRows       int
	runTimeout    float64
	runRecordPath string
	runEnv        []string
	runWorkingDir string
)

func init() {
	runCmd.Flags().StringVar(&runGoldenPath, "golden", "", "golden file to verify against")
	runCmd.Flags().BoolVar(&runUpdate, "update", false, "update golden file instead of verifying")
	runCmd.Flags().IntVar(&runCols, "cols", 108, "terminal width in columns")
	runCmd.Flags().IntVar(&runRows, "rows", 24, "terminal height in rows")
	runCmd.Flags().Float64Var(&runTimeout, "timeout", 30, "timeout in seconds for Wait commands")
	runCmd.Flags().StringVar(&runRecordPath, "record", "", "record session to .cast file")
	runCmd.Flags().StringArrayVar(&runEnv, "env", nil, "environment variables (key=value, repeatable)")
	runCmd.Flags().StringVar(&runWorkingDir, "working-directory", "", "working directory for the child process")

	rootCmd.AddCommand(runCmd)

	carapace.Gen(runCmd).FlagCompletion(carapace.ActionMap{
		"golden": carapace.ActionFiles(),
		"record": carapace.ActionFiles(".cast"),
	})
	carapace.Gen(runCmd).PositionalCompletion(
		carapace.ActionFiles(".tape"),
	)
}
