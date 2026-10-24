package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tab-dance/tabdance-test/internal/service"
)

// handleRunTape is the MCP equivalent of the CLI `run` command; all
// semantics (Set Shell honoring, bash OSC integration, typing speed,
// timeout resolution, exit status, strict mode) live in service.RunTape.
func (s *Server) handleRunTape(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tapePath, err := req.RequireString("tape_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	timeout := time.Duration(req.GetInt("timeout", 30)) * time.Second

	outcome, err := s.svc.RunTape(service.RunTapeOptions{
		TapePath:          tapePath,
		GoldenPath:        req.GetString("golden_path", ""),
		Update:            req.GetBool("update", false),
		Strict:            req.GetBool("strict", false),
		Timeout:           timeout,
		PromptIntegration: true,
	})
	if err != nil {
		msg := err.Error()
		if outcome != nil && outcome.Exited {
			// Surface the child's termination alongside the failure.
			if outcome.ExitSignal != "" {
				msg += fmt.Sprintf("\n(child killed by %s)", outcome.ExitSignal)
			} else {
				msg += fmt.Sprintf("\n(child exited with code %d)", outcome.ExitCode)
			}
		}
		return mcp.NewToolResultError(msg), nil
	}

	if !outcome.Verified {
		return mcp.NewToolResultText(fmt.Sprintf("tape %q executed", tapePath)), nil
	}
	if outcome.Updated {
		return mcp.NewToolResultText(fmt.Sprintf("golden file %q updated", outcome.GoldenPath)), nil
	}
	if !outcome.Match {
		return mcp.NewToolResultError(fmt.Sprintf("output does not match golden file %s\n%s", outcome.GoldenPath, outcome.Diff)), nil
	}

	out := map[string]any{"match": true}
	if outcome.Exited {
		out["exit_code"] = outcome.ExitCode
		out["exit_signal"] = outcome.ExitSignal
	}
	data, _ := json.Marshal(out)
	return mcp.NewToolResultText(string(data)), nil
}
