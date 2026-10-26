package service

import (
	"fmt"
	"os"
	"time"

	"github.com/tab-dance/tabdance-test/internal/golden"
	"github.com/tab-dance/tabdance-test/internal/tape"
	"github.com/tab-dance/tabdance-test/internal/terminal"
)

// RunTapeOptions configures a fresh-terminal tape run shared by the CLI
// `run` command and the MCP run_tape tool.
type RunTapeOptions struct {
	TapePath          string
	GoldenPath        string
	Update            bool
	Strict            bool
	Timeout           time.Duration // 0 → tape WaitTimeout, else 30s
	Cols, Rows        int           // 0 → tape Set Width/Height, else 108/24
	Scrollback        int           // 0 → 10000
	Env               map[string]string
	WorkingDirectory  string
	RecordPath        string // asciicast recording for the run terminal
	PromptIntegration bool   // inject OSC 133 integration for bash tapes
}

// RunTapeOutcome describes a completed run beyond the error.
type RunTapeOutcome struct {
	Progress   []string // quiet lines: "PASS (exit=0)", "UPDATED <path>"
	Executed   bool     // tape commands ran (true once executor finished)
	Verified   bool     // golden verification or update happened
	Match      bool     // verification result (only when Verified)
	Diff       string   // on mismatch
	Updated    bool     // golden file was updated instead of verified
	GoldenPath string   // verified (or updated) golden file
	Exited     bool
	ExitCode   int
	ExitSignal string
}

// RunTape honors the tape's Set Shell/Width/Height/WaitTimeout/TypingSpeed
// settings, plays the tape in a fresh terminal, optionally verifies (or
// updates) a golden file, and always tears the terminal down.
func (s *Service) RunTape(o RunTapeOptions) (*RunTapeOutcome, error) {
	out := &RunTapeOutcome{}

	if o.Strict {
		if err := tape.CheckStrict(o.TapePath); err != nil {
			return out, err
		}
	}

	settings, err := tape.ParseSettings(o.TapePath)
	if err != nil {
		return out, fmt.Errorf("parse tape settings: %w", err)
	}

	cols := o.Cols
	if cols <= 0 {
		cols = 108
	}
	rows := o.Rows
	if rows <= 0 {
		rows = 24
	}
	if settings.Width > 0 {
		cols = settings.Width
	}
	if settings.Height > 0 {
		rows = settings.Height
	}
	scrollback := o.Scrollback
	if scrollback <= 0 {
		scrollback = 10000
	}

	shellCommand := "sh"
	if settings.Shell != "" {
		shellCommand = tape.ShellToCommand(settings.Shell)
	}
	env := map[string]string{}
	for k, v := range o.Env {
		env[k] = v
	}
	if shellCommand == "bash --norc --noprofile" && o.PromptIntegration {
		// OSC 133 integration via the environment: no rc files, escape
		// bytes only, so rendered screens and goldens are unaffected.
		env["PROMPT_COMMAND"] = `printf '\033]133;D;%s\007\033]133;A\007' "$?"`
		env["PS0"] = `\033]133;C\007`
	}

	timeout := o.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	if settings.WaitTimeout > 0 {
		timeout = settings.WaitTimeout
	}

	created, err := s.Exec(terminal.CreateConfig{
		Cols:             uint16(cols),
		Rows:             uint16(rows),
		Scrollback:       uint(scrollback),
		Command:          shellCommand,
		Env:              env,
		WorkingDirectory: o.WorkingDirectory,
	}, o.RecordPath, 0)
	if err != nil {
		return out, err
	}
	mt, err := s.terminals.Get(created.TerminalID)
	if err != nil {
		return out, err
	}
	defer s.Delete(created.TerminalID)

	executor := tape.NewTapeExecutor(mt, timeout)
	if settings.TypingSpeed > 0 {
		executor.SetTypingSpeed(settings.TypingSpeed)
	}

	// Checkpoint collector: every anchor (Expect, Wait+Screen) records a
	// labeled golden section; the final screen becomes the last section.
	type section struct {
		label   string
		content string
	}
	var sections []section
	if o.GoldenPath != "" {
		executor.SetCheckpoints(func(label string) {
			mt.Lock()
			content, err := golden.Capture(mt.Terminal, "plain", true, true)
			mt.Unlock()
			if err != nil {
				return
			}
			sections = append(sections, section{label: label, content: content})
		})
	}

	if err := executor.Execute(o.TapePath); err != nil {
		return out, fmt.Errorf("execute tape: %w", err)
	}
	out.Executed = true

	mt.StopRecording()

	out.Exited, out.ExitCode, out.ExitSignal = mt.ExitState()

	if o.GoldenPath == "" {
		return out, nil
	}
	out.GoldenPath = o.GoldenPath
	out.Verified = true
	out.Updated = o.Update

	// Final section: end-of-tape screen state.
	mt.Lock()
	finalContent, err := golden.Capture(mt.Terminal, "plain", true, true)
	mt.Unlock()
	if err != nil {
		return out, err
	}
	sections = append(sections, section{label: "final", content: finalContent})

	labels := make([]string, len(sections))
	contents := make([]string, len(sections))
	for i, sec := range sections {
		labels[i] = sec.label
		contents[i] = sec.content
	}

	if o.Update {
		if err := golden.SaveSections(o.GoldenPath, labels, contents); err != nil {
			return out, err
		}
		out.Progress = append(out.Progress, fmt.Sprintf("UPDATED %s", o.GoldenPath))
		return out, nil
	}

	expectedLabels, expectedContents, err := golden.LoadSections(o.GoldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return out, fmt.Errorf("FAIL: golden file %q does not exist", o.GoldenPath)
		}
		return out, fmt.Errorf("read golden: %w", err)
	}

	actual := golden.SectionsText(labels, contents)
	expected := golden.SectionsText(expectedLabels, expectedContents)
	if len(expectedLabels) != len(labels) {
		return out, fmt.Errorf("FAIL: golden has %d sections, replay produced %d",
			len(expectedLabels), len(labels))
	}
	for i := range labels {
		if i < len(expectedLabels) && expectedLabels[i] != labels[i] {
			return out, fmt.Errorf("FAIL: section %d label differs: golden %q, anchor %q",
				i+1, expectedLabels[i], labels[i])
		}
	}
	if expected != actual {
		for i := range contents {
			if contents[i] != expectedContents[i] {
				diff, derr := golden.DiffText(expectedContents[i], contents[i], o.GoldenPath)
				if derr == nil {
					return out, fmt.Errorf("FAIL: section %d (%s) differs\n%s", i+1, labels[i], diff)
				}
			}
		}
		if diff, derr := golden.DiffText(expected, actual, o.GoldenPath); derr == nil {
			return out, fmt.Errorf("FAIL\n%s", diff)
		}
	}
	out.Match = true

	msg := "PASS"
	if out.Exited {
		if out.ExitSignal != "" {
			msg += fmt.Sprintf(" (child killed by %s)", out.ExitSignal)
		} else {
			msg += fmt.Sprintf(" (exit=%d)", out.ExitCode)
		}
	}
	out.Progress = append(out.Progress, msg)
	return out, nil
}

func basename(p string) string {
	if idx := indexAnyLast(p, "/\\"); idx >= 0 {
		return p[idx+1:]
	}
	return p
}

// indexAnyLast mirrors strings.LastIndexAny without importing more than
// needed here.
func indexAnyLast(s, chars string) int {
	last := -1
	for i := 0; i < len(s); i++ {
		for j := 0; j < len(chars); j++ {
			if s[i] == chars[j] {
				last = i
				break
			}
		}
	}
	return last
}
