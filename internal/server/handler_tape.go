package server

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handlePlayTape(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return handleParamError(err)
	}
	tapePath, err := req.RequireString("tape_path")
	if err != nil {
		return handleParamError(err)
	}
	timeout := time.Duration(req.GetInt("timeout", 30)) * time.Second

	if err := s.svc.PlayTape(id, tapePath, timeout); err != nil {
		return handleParamError(err)
	}
	return mcp.NewToolResultText(fmt.Sprintf("tape %q executed", tapePath)), nil
}
