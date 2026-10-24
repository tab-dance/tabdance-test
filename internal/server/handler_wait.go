package server

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
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
		return mcp.NewToolResultError(fmt.Sprintf("pattern %q not found within %v", pattern, timeout)), nil
	}

	return mcp.NewToolResultText(result.ToJSON()), nil
}
