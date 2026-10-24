package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
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
			_, err = runTape(args[0], runGoldenPath, runUpdate)
		}
		if err != nil && runGoldenPath != "" && !runUpdate {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		return err
	},
}

// runTape plays a single tape and verifies (or updates) its golden file.
// A verification mismatch is returned as an error whose message is the
// "FAIL\n<diff>" report; quiet output (PASS/UPDATED, with exit status)
// is returned via progress so callers render it where they like.
func runTape(tapePath, goldenPath string, update bool) (progress []string, err error) {
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
		return progress, fmt.Errorf("parse tape settings: %w", err)
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
		return progress, fmt.Errorf("create terminal: %w", err)
	}
	defer mt.Close()

	if runRecordPath != "" {
		if err := mt.StartRecording(runRecordPath, 2.0); err != nil {
			return progress, fmt.Errorf("start recording: %w", err)
		}
	}

	executor := tape.NewTapeExecutor(mt, timeout)
	if settings.TypingSpeed > 0 {
		executor.SetTypingSpeed(settings.TypingSpeed)
	}
	if err := executor.Execute(tapePath); err != nil {
		return progress, fmt.Errorf("execute tape: %w", err)
	}

	mt.StopRecording()

	if goldenPath == "" {
		return progress, nil
	}

	if update {
		mt.Lock()
		err := golden.UpdateGolden(mt.Terminal, goldenPath, "plain", true, true)
		mt.Unlock()
		if err != nil {
			return progress, err
		}
		return append(progress, fmt.Sprintf("UPDATED %s", goldenPath)), nil
	}

	mt.Lock()
	result, err := golden.VerifyOutput(mt.Terminal, goldenPath, "plain", true, true)
	mt.Unlock()
	if err != nil {
		return progress, err
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
		return append(progress, msg), nil
	}

	failMsg := fmt.Sprintf("FAIL\n%s", result.Diff)
	if _, code, signal := mt.ExitState(); code != 0 || signal != "" {
		if signal != "" {
			failMsg += fmt.Sprintf("\n(child killed by %s)", signal)
		} else {
			failMsg += fmt.Sprintf("\n(child exited with code %d)", code)
		}
	}
	return progress, fmt.Errorf("%s", failMsg)
}

// dirResult carries one tape's outcome in directory mode. Input order is
// preserved (the slice is sorted by tape name whatever the concurrency).
type dirResult struct {
	Tape     string
	Progress []string // PASS/UPDATED lines, in order
	Err      error    // strict-check or tape failure; nil = passed
	Skipped  bool
}

func (r dirResult) Base() string { return filepath.Base(r.Tape) }

// runDirectoryResults plays every *.tape in the directory (against the
// matching .golden unless --update, which skips tapes without one),
// running up to jobs concurrently. No printing, no exits: callers decide
// how to render and whether to abort.
func runDirectoryResults(dir string, jobs int, recordPath string) ([]dirResult, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.tape"))
	if err != nil {
		return nil, fmt.Errorf("glob %s: %w", dir, err)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no tapes found in %s", dir)
	}
	sort.Strings(matches)

	if jobs < 1 {
		jobs = 1
	}
	if jobs > len(matches) {
		jobs = len(matches)
	}
	if jobs > 1 && recordPath != "" {
		return nil, fmt.Errorf("--record cannot be used with --jobs > 1 (recordings would interleave in one file)")
	}

	runOne := func(tapePath string) dirResult {
		res := dirResult{Tape: tapePath}
		if !runUpdate {
			goldenPath := strings.TrimSuffix(tapePath, ".tape") + ".golden"
			if _, err := os.Stat(goldenPath); err != nil {
				res.Skipped = true
				return res
			}
		}
		if err := runStrictCheck(tapePath); err != nil {
			res.Err = err
			return res
		}
		goldenPath := strings.TrimSuffix(tapePath, ".tape") + ".golden"
		res.Progress, res.Err = runTape(tapePath, goldenPath, runUpdate)
		return res
	}

	results := make([]dirResult, len(matches))
	if jobs == 1 {
		for i, tapePath := range matches {
			results[i] = runOne(tapePath)
		}
		return results, nil
	}

	// Worker pool: each tape owns an isolated terminal, so parallelism
	// only contends on CPU and the machine's PTY limits, not state.
	queue := make(chan int)
	var mu sync.Mutex
	var wg sync.WaitGroup
	for worker := 0; worker < jobs; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range queue {
				res := runOne(matches[i])
				mu.Lock()
				results[i] = res
				mu.Unlock()
			}
		}()
	}
	for i := range matches {
		queue <- i
	}
	close(queue)
	wg.Wait()

	sort.Slice(results, func(i, j int) bool { return results[i].Tape < results[j].Tape })
	return results, nil
}

// runDirectory renders the aggregated results and exits nonzero when any
// tape failed, keeping the CI contract of the serial runner.
func runDirectory(dir string) error {
	results, err := runDirectoryResults(dir, runJobs, runRecordPath)
	if err != nil {
		return err
	}

	var passed, failed, skipped []string
	for _, r := range results {
		switch {
		case r.Skipped:
			skipped = append(skipped, r.Base())
		case r.Err != nil:
			failed = append(failed, r.Base())
		default:
			passed = append(passed, r.Base())
		}
	}

	for _, r := range results {
		if r.Skipped {
			continue
		}
		fmt.Fprintf(os.Stderr, "== %s\n", r.Base())
		for _, line := range r.Progress {
			fmt.Fprintln(os.Stderr, line)
		}
		if r.Err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", r.Err)
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
	runJobs       int
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
	runCmd.Flags().StringVar(&runRecordPath, "record", "", "record session to .cast file (one tape only)")
	runCmd.Flags().StringArrayVar(&runEnv, "env", nil, "environment variables (key=value, repeatable)")
	runCmd.Flags().StringVar(&runWorkingDir, "working-directory", "", "working directory for the child process")
	runCmd.Flags().BoolVar(&runJsonOutput, "json", false, "when running a directory, print the aggregated result as JSON")
	runCmd.Flags().BoolVar(&runStrict, "strict", false, "reject tapes containing Sleep verbs (determinism guard)")
	runCmd.Flags().StringVar(&runTerm, "term", "", "override the TERM environment variable for the child process")
	runCmd.Flags().IntVar(&runJobs, "jobs", 1, "number of tapes to run concurrently (directory mode only)")

	rootCmd.AddCommand(runCmd)

	carapace.Gen(runCmd).FlagCompletion(carapace.ActionMap{
		"golden": carapace.ActionFiles(),
		"record": carapace.ActionFiles(".cast"),
	})
	carapace.Gen(runCmd).PositionalCompletion(
		carapace.ActionFiles(".tape"),
	)
}
