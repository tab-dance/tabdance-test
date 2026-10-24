package server

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tab-dance/tabdance-test/internal/inspect"
	"github.com/tab-dance/tabdance-test/internal/terminal"
	"github.com/tab-dance/tabdance-test/internal/wait"
)

func (s *Server) handleWaitFor(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	pattern, err := req.RequireString("pattern")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	scope := req.GetString("scope", "screen")
	timeout := time.Duration(req.GetInt("timeout", 10)) * time.Second
	format := req.GetString("format", "plain")
	trim := req.GetBool("trim", true)
	unwrap := req.GetBool("unwrap", true)

	mt, err := s.terminals.Get(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result, err := wait.WaitFor(mt, mt.Terminal, wait.WaitForOptions{
		Pattern: pattern,
		Scope:   scope,
		Timeout: timeout,
		Format:  format,
		Trim:    trim,
		Unwrap:  unwrap,
	})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if !result.Found {
		return mcp.NewToolResultError(waitFailureDetail(pattern, timeout, mt, format, trim, unwrap)), nil
	}

	return mcp.NewToolResultText(result.ToJSON()), nil
}

// waitFailureDetail augments a wait timeout with the rendered screen and
// the last PTY traffic so failures are diagnosable without attach.
func waitFailureDetail(pattern string, timeout time.Duration, mt *terminal.ManagedTerminal, format string, trim, unwrap bool) string {
	mt.Lock()
	view, err := inspect.GetView(mt.Terminal, format, trim, unwrap)
	mt.Unlock()

	msg := fmt.Sprintf("pattern %q not found within %v", pattern, timeout)
	if err != nil {
		return fmt.Sprintf("%s\n(could not render screen: %v)", msg, err)
	}
	recent := mt.RecentOutput(2048)
	detail := msg + "\n--- screen at timeout ---\n" + view.Content
	if len(recent) > 0 {
		detail += fmt.Sprintf("\n--- recent output (last %d bytes) ---\n%s", len(recent), wait.Repr(recent))
	}
	return detail
}
