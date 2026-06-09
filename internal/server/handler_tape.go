package server

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tab-dance/tabdance-test/internal/tape"
)

func (s *Server) handlePlayTape(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	tapePath, err := req.RequireString("tape_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	timeout := time.Duration(req.GetInt("timeout", 30)) * time.Second

	mt, err := s.terminals.Get(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	executor := tape.NewTapeExecutor(mt, timeout)
	if err := executor.Execute(tapePath); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("tape %q executed", tapePath)), nil
}
