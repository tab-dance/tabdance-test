package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/tab-dance/tabdance-test/internal/golden"
	"github.com/tab-dance/tabdance-test/internal/tape"
	"github.com/tab-dance/tabdance-test/internal/terminal"
)

func runRunInProcess(tapePath string, opts runOptions) error {
	env := make(map[string]string)
	for _, e := range opts.env {
		k, v, _ := parseKeyValue(e)
		env[k] = v
	}

	mt, err := terminal.NewManagedTerminal("run_1", terminal.CreateConfig{
		Cols:             uint16(opts.cols),
		Rows:             uint16(opts.rows),
		Scrollback:       10000,
		Command:          "", // tape provides the command via Set Shell
		Env:              env,
		WorkingDirectory: opts.workingDir,
	})
	if err != nil {
		return fmt.Errorf("create terminal: %w", err)
	}
	defer mt.Close()

	if opts.recordPath != "" {
		if err := mt.StartRecording(opts.recordPath, 2.0); err != nil {
			return fmt.Errorf("start recording: %w", err)
		}
	}

	executor := tape.NewTapeExecutor(mt, time.Duration(opts.timeout*float64(time.Second)))
	if err := executor.Execute(tapePath); err != nil {
		return fmt.Errorf("execute tape: %w", err)
	}

	if mt.Recorder != nil {
		mt.StopRecording()
	}

	if opts.goldenPath != "" {
		if opts.update {
			mt.Lock()
			err := golden.UpdateGolden(mt.Terminal, opts.goldenPath, "plain", true, true)
			mt.Unlock()
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "UPDATED %s\n", opts.goldenPath)
			return nil
		}

		mt.Lock()
		result, err := golden.VerifyOutput(mt.Terminal, opts.goldenPath, "plain", true, true)
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
}
