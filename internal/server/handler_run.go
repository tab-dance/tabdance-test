package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tab-dance/tabdance-test/internal/golden"
	"github.com/tab-dance/tabdance-test/internal/tape"
	"github.com/tab-dance/tabdance-test/internal/terminal"
)

// handleRunTape is the MCP equivalent of the CLI `run` command: it creates a
// fresh terminal honoring the tape's Set Shell/Width/Height/WaitTimeout/
// TypingSpeed settings, plays the tape, optionally verifies (or updates) a
// golden file, and tears the terminal down again.
func (s *Server) handleRunTape(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tapePath, err := req.RequireString("tape_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	timeout := time.Duration(req.GetInt("timeout", 30)) * time.Second
	goldenPath := req.GetString("golden_path", "")
	update := req.GetBool("update", false)
	strict := req.GetBool("strict", false)

	settings, err := tape.ParseSettings(tapePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("parse tape settings: %v", err)), nil
	}
	if strict {
		if err := tape.CheckStrict(tapePath); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	}

	cols := 108
	rows := 24
	if settings.Width > 0 {
		cols = settings.Width
	}
	if settings.Height > 0 {
		rows = settings.Height
	}
	if settings.WaitTimeout > 0 {
		timeout = settings.WaitTimeout
	}
	shellCommand := "sh"
	if settings.Shell != "" {
		shellCommand = tape.ShellToCommand(settings.Shell)
	}

	created, err := s.svc.Exec(terminal.CreateConfig{
		Cols:       uint16(cols),
		Rows:       uint16(rows),
		Scrollback: 10000,
		Command:    shellCommand,
	}, "", 0)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer s.svc.Delete(created.TerminalID)
	mt, err := s.terminals.Get(created.TerminalID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	executor := tape.NewTapeExecutor(mt, timeout)
	if settings.TypingSpeed > 0 {
		executor.SetTypingSpeed(settings.TypingSpeed)
	}
	if err := executor.Execute(tapePath); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("execute tape: %v", err)), nil
	}

	if goldenPath == "" {
		return mcp.NewToolResultText(fmt.Sprintf("tape %q executed", tapePath)), nil
	}

	mt.Lock()
	var result *golden.VerifyResult
	if update {
		err = golden.UpdateGolden(mt.Terminal, goldenPath, "plain", true, true)
	} else {
		result, err = golden.VerifyOutput(mt.Terminal, goldenPath, "plain", true, true)
	}
	mt.Unlock()

	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if update {
		return mcp.NewToolResultText(fmt.Sprintf("golden file %q updated", goldenPath)), nil
	}

	if !result.Match {
		return mcp.NewToolResultError(runTapeFailure(mt, result.Diff)), nil
	}
	return mcp.NewToolResultText(runTapeResultJSON(mt, result)), nil
}

// runTapeResultJSON marshals the verification result, adding the child's
// exit status when it has already terminated.
func runTapeResultJSON(mt *terminal.ManagedTerminal, result *golden.VerifyResult) string {
	out := map[string]any{"match": result.Match}
	if exited, code, signal := mt.ExitState(); exited {
		out["exit_code"] = code
		out["exit_signal"] = signal
	}
	data, _ := json.Marshal(out)
	return string(data)
}

// runTapeFailure renders a verification failure with child exit status.
func runTapeFailure(mt *terminal.ManagedTerminal, diff string) string {
	msg := fmt.Sprintf("FAIL\n%s", diff)
	if _, code, signal := mt.ExitState(); code != 0 || signal != "" {
		if signal != "" {
			msg += fmt.Sprintf("\n(child killed by signal %s)", signal)
		} else {
			msg += fmt.Sprintf("\n(child exited with code %d)", code)
		}
	}
	return msg
}
