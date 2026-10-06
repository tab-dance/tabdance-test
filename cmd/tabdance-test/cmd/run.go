package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/tab-dance/tabdance-test/internal/golden"
	"github.com/tab-dance/tabdance-test/internal/tape"
	"github.com/tab-dance/tabdance-test/internal/terminal"
)

var runCmd = &cobra.Command{
	Use:   "run <tape-or-directory>",
	Short: "Play tape(s), verify against golden, exit 0/1 (CI-friendly)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := os.Stat(args[0])
		if err != nil {
			return fmt.Errorf("stat %s: %w", args[0], err)
		}
		if info.IsDir() {
			if runGoldenPath != "" {
				return fmt.Errorf("--golden cannot be used with a directory")
			}
			return runDirectory(args[0])
		}

		err = runTape(args[0], runGoldenPath, runUpdate)
		if err != nil && runGoldenPath != "" && !runUpdate {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		return err
	},
}

// runTape plays a single tape and verifies (or updates) its golden file.
// a verification mismatch is returned as an error whose message is the
// "FAIL\n<diff>" report.
func runTape(tapePath, goldenPath string, update bool) error {
	env := make(map[string]string)
	for _, e := range runEnv {
		k, v, _ := parseKeyValue(e)
		env[k] = v
	}

	settings, err := tape.ParseSettings(tapePath)
	if err != nil {
		return fmt.Errorf("parse tape settings: %w", err)
	}

	cols := runCols
	rows := runRows
	if settings.Width > 0 {
		cols = settings.Width
	}
	if settings.Height > 0 {
		rows = settings.Height
	}

	shellCommand := "sh"
	if settings.Shell != "" {
		shellCommand = tape.ShellToCommand(settings.Shell)
	}

	timeout := time.Duration(runTimeout * float64(time.Second))
	if settings.WaitTimeout > 0 {
		timeout = settings.WaitTimeout
	}

	mt, err := terminal.NewManagedTerminal(fmt.Sprintf("run_%s", strings.TrimSuffix(filepath.Base(tapePath), ".tape")), terminal.CreateConfig{
		Cols:             uint16(cols),
		Rows:             uint16(rows),
		Scrollback:       10000,
		Command:          shellCommand,
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

	executor := tape.NewTapeExecutor(mt, timeout)
	if settings.TypingSpeed > 0 {
		executor.SetTypingSpeed(settings.TypingSpeed)
	}
	if runPreTape != "" {
		if err := executor.Execute(runPreTape); err != nil {
			return fmt.Errorf("execute setup tape: %w", err)
		}
	}
	if err := executor.Execute(tapePath); err != nil {
		return fmt.Errorf("execute tape: %w", err)
	}

	if mt.Recorder != nil {
		mt.StopRecording()
	}

	if goldenPath == "" {
		return nil
	}

	if update {
		mt.Lock()
		err := golden.UpdateGolden(mt.Terminal, goldenPath, "plain", true, true)
		mt.Unlock()
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "UPDATED %s\n", goldenPath)
		return nil
	}

	mt.Lock()
	result, err := golden.VerifyOutput(mt.Terminal, goldenPath, "plain", true, true)
	mt.Unlock()
	if err != nil {
		return err
	}

	if result.Match {
		fmt.Fprintf(os.Stderr, "PASS\n")
		return nil
	}

	return fmt.Errorf("FAIL\n%s", result.Diff)
}

// runDirectory plays every *.tape in the directory against the matching
// .golden file (tapes without one are skipped unless --update) and reports
// an aggregated result so a single run shows all broken tapes.
func runDirectory(dir string) error {
	matches, err := filepath.Glob(filepath.Join(dir, "*.tape"))
	if err != nil {
		return fmt.Errorf("glob %s: %w", dir, err)
	}
	if len(matches) == 0 {
		return fmt.Errorf("no tapes found in %s", dir)
	}
	sort.Strings(matches)

	var failed, skipped []string
	for _, tapePath := range matches {
		goldenPath := strings.TrimSuffix(tapePath, ".tape") + ".golden"
		if !runUpdate {
			if _, err := os.Stat(goldenPath); err != nil {
				skipped = append(skipped, filepath.Base(tapePath))
				continue
			}
		}
		fmt.Fprintf(os.Stderr, "== %s\n", filepath.Base(tapePath))
		if err := runTape(tapePath, goldenPath, runUpdate); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			failed = append(failed, filepath.Base(tapePath))
		}
	}

	if len(skipped) > 0 {
		fmt.Fprintf(os.Stderr, "SKIPPED (no golden yet): %s\n", strings.Join(skipped, ", "))
	}
	if len(failed) > 0 {
		fmt.Fprintf(os.Stderr, "FAILED: %s\n", strings.Join(failed, ", "))
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "ALL PASS\n")
	return nil
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
	runPreTape    string
)

func init() {
	runCmd.Flags().StringVar(&runGoldenPath, "golden", "", "golden file to verify against (derived from the tape name when running a directory)")
	runCmd.Flags().BoolVar(&runUpdate, "update", false, "update golden file instead of verifying")
	runCmd.Flags().IntVar(&runCols, "cols", 108, "terminal width in columns")
	runCmd.Flags().IntVar(&runRows, "rows", 24, "terminal height in rows")
	runCmd.Flags().Float64Var(&runTimeout, "timeout", 30, "timeout in seconds for Wait commands")
	runCmd.Flags().StringVar(&runRecordPath, "record", "", "record session to .cast file")
	runCmd.Flags().StringArrayVar(&runEnv, "env", nil, "environment variables (key=value, repeatable)")
	runCmd.Flags().StringVar(&runWorkingDir, "working-directory", "", "working directory for the child process")
	runCmd.Flags().StringVar(&runPreTape, "pre", "", "tape to execute before the main tape (e.g. shell setup)")

	rootCmd.AddCommand(runCmd)

	carapace.Gen(runCmd).FlagCompletion(carapace.ActionMap{
		"golden": carapace.ActionFiles(),
		"record": carapace.ActionFiles(".cast"),
	})
	carapace.Gen(runCmd).PositionalCompletion(
		carapace.ActionFiles(".tape"),
	)
}
