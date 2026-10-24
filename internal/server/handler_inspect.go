package server

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handleGetView(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return handleParamError(err)
	}
	format := req.GetString("format", "plain")
	trim := req.GetBool("trim", true)
	unwrap := req.GetBool("unwrap", true)

	result, err := s.svc.View(id, format, trim, unwrap)
	if err != nil {
		return handleParamError(err)
	}
	return mcp.NewToolResultText(result.ToJSON()), nil
}

func (s *Server) handleGetCursor(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return handleParamError(err)
	}
	result, err := s.svc.Cursor(id)
	if err != nil {
		return handleParamError(err)
	}
	return mcp.NewToolResultText(result.ToJSON()), nil
}

func (s *Server) handleGetSize(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return handleParamError(err)
	}
	result, err := s.svc.Size(id)
	if err != nil {
		return handleParamError(err)
	}
	return mcp.NewToolResultText(result.ToJSON()), nil
}

func (s *Server) handleGetCell(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return handleParamError(err)
	}
	x := req.GetInt("x", 0)
	y := req.GetInt("y", 0)
	result, err := s.svc.Cell(id, x, y)
	if err != nil {
		return handleParamError(err)
	}
	return mcp.NewToolResultText(result.ToJSON()), nil
}

// handleGetOutput returns the terminal's recent raw PTY output, with
// control bytes rendered visibly, for diagnosing failures without attach.
func (s *Server) handleGetOutput(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return handleParamError(err)
	}
	n := req.GetInt("bytes", 4096)

	out, err := s.svc.RecentOutput(id, n)
	if err != nil {
		return handleParamError(err)
	}
	return mcp.NewToolResultText(out), nil
}
