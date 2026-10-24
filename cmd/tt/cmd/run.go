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
	"github.com/tab-dance/tabdance-test/internal/service"
	"github.com/tab-dance/tabdance-test/internal/tape"
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

		if runStrict {
			if err := runStrictCheck(args[0]); err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
		}
		progress, err := runTape(args[0], runGoldenPath, runUpdate)
		for _, line := range progress {
			fmt.Fprintln(os.Stderr, line)
		}
		if err != nil && runGoldenPath != "" && !runUpdate {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		return err
	},
}

// runTape plays a single tape, translating package-level run flags into
// the transport-neutral service operation. A verification mismatch is
// returned as an error whose message is the "FAIL<diff>" report; PASS/
// UPDATED lines (with exit status) come back via progress.
func runTape(tapePath, goldenPath string, update bool) ([]string, error) {
	env := make(map[string]string)
	for _, e := range runEnv {
		k, v, _ := parseKeyValue(e)
		env[k] = v
	}
	if runTerm != "" {
		env["TERM"] = runTerm
	}

	outcome, err := service.New().RunTape(service.RunTapeOptions{
		TapePath:          tapePath,
		GoldenPath:        goldenPath,
		Update:            update,
		Strict:            runStrict,
		Timeout:           time.Duration(runTimeout * float64(time.Second)),
		Cols:              runCols,
		Rows:              runRows,
		Scrollback:        runScrollback,
		Env:               env,
		WorkingDirectory:  runWorkingDir,
		RecordPath:        runRecordPath,
		PromptIntegration: !runNoPromptIntegration,
	})
	return outcome.Progress, err
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
	runGoldenPath          string
	runUpdate              bool
	runCols                int
	runRows                int
	runScrollback          int
	runTimeout             float64
	runRecordPath          string
	runEnv                 []string
	runWorkingDir          string
	runJsonOutput          bool
	runStrict              bool
	runTerm                string
	runJobs                int
	runNoPromptIntegration bool
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
	runCmd.Flags().BoolVar(&runNoPromptIntegration, "no-prompt-integration", false, "do not inject OSC 133 shell integration for bash tapes")

	rootCmd.AddCommand(runCmd)

	carapace.Gen(runCmd).FlagCompletion(carapace.ActionMap{
		"golden": carapace.ActionFiles(),
		"record": carapace.ActionFiles(".cast"),
	})
	carapace.Gen(runCmd).PositionalCompletion(
		carapace.ActionFiles(".tape"),
	)
}
