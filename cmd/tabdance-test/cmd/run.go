package cmd

import (
	"encoding/json"
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

		err = runStrictCheck(args[0])
		if err == nil {
			err = runTape(args[0], runGoldenPath, runUpdate)
		}
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
	if runTerm != "" {
		env["TERM"] = runTerm
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
		Scrollback:       uint(runScrollback),
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
	if err := executor.Execute(tapePath); err != nil {
		return fmt.Errorf("execute tape: %w", err)
	}

	mt.StopRecording()

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
		msg := "PASS"
		if exited, code, signal := mt.ExitState(); exited {
			if signal != "" {
				msg += fmt.Sprintf(" (child killed by %s)", signal)
			} else {
				msg += fmt.Sprintf(" (exit=%d)", code)
			}
		}
		fmt.Fprintf(os.Stderr, "%s\n", msg)
		return nil
	}

	failMsg := fmt.Sprintf("FAIL\n%s", result.Diff)
	if _, code, signal := mt.ExitState(); code != 0 || signal != "" {
		if signal != "" {
			failMsg += fmt.Sprintf("\n(child killed by %s)", signal)
		} else {
			failMsg += fmt.Sprintf("\n(child exited with code %d)", code)
		}
	}
	return fmt.Errorf("%s", failMsg)
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

	var failed, skipped, passed []string
	for _, tapePath := range matches {
		goldenPath := strings.TrimSuffix(tapePath, ".tape") + ".golden"
		if !runUpdate {
			if _, err := os.Stat(goldenPath); err != nil {
				skipped = append(skipped, filepath.Base(tapePath))
				continue
			}
		}
		fmt.Fprintf(os.Stderr, "== %s\n", filepath.Base(tapePath))
		if err := runStrictCheck(tapePath); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			failed = append(failed, filepath.Base(tapePath))
			continue
		}
		if err := runTape(tapePath, goldenPath, runUpdate); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			failed = append(failed, filepath.Base(tapePath))
		} else {
			passed = append(passed, filepath.Base(tapePath))
		}
	}

	if runJsonOutput {
		data, _ := json.Marshal(map[string]any{
			"passed":  passed,
			"failed":  failed,
			"skipped": skipped,
		})
		fmt.Println(string(data))
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
	runScrollback int
	runTimeout    float64
	runRecordPath string
	runEnv        []string
	runWorkingDir string
	runJsonOutput bool
	runStrict     bool
	runTerm       string
)

func runStrictCheck(tapePath string) error {
	if !runStrict {
		return nil
	}
	return tape.CheckStrict(tapePath)
}

func init() {
	runCmd.Flags().StringVar(&runGoldenPath, "golden", "", "golden file to verify against (derived from the tape name when running a directory)")
	runCmd.Flags().BoolVar(&runUpdate, "update", false, "update golden file instead of verifying")
	runCmd.Flags().IntVar(&runCols, "cols", 108, "terminal width in columns")
	runCmd.Flags().IntVar(&runRows, "rows", 24, "terminal height in rows")
	runCmd.Flags().IntVar(&runScrollback, "scrollback", 10000, "max scrollback lines")
	runCmd.Flags().Float64Var(&runTimeout, "timeout", 30, "timeout in seconds for Wait commands")
	runCmd.Flags().StringVar(&runRecordPath, "record", "", "record session to .cast file")
	runCmd.Flags().StringArrayVar(&runEnv, "env", nil, "environment variables (key=value, repeatable)")
	runCmd.Flags().StringVar(&runWorkingDir, "working-directory", "", "working directory for the child process")
	runCmd.Flags().BoolVar(&runJsonOutput, "json", false, "when running a directory, print the aggregated result as JSON")
	runCmd.Flags().BoolVar(&runStrict, "strict", false, "reject tapes containing Sleep verbs (determinism guard)")
	runCmd.Flags().StringVar(&runTerm, "term", "", "override the TERM environment variable for the child process")

	rootCmd.AddCommand(runCmd)

	carapace.Gen(runCmd).FlagCompletion(carapace.ActionMap{
		"golden": carapace.ActionFiles(),
		"record": carapace.ActionFiles(".cast"),
	})
	carapace.Gen(runCmd).PositionalCompletion(
		carapace.ActionFiles(".tape"),
	)
}
