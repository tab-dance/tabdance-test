package server

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tab-dance/tabdance-test/internal/golden"
)

func (s *Server) handleVerifyOutput(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	goldenPath, err := req.RequireString("golden_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	format := req.GetString("format", "plain")
	trim := req.GetBool("trim", true)
	unwrap := req.GetBool("unwrap", true)

	mt, err := s.terminals.Get(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	mt.Lock()
	defer mt.Unlock()

	result, err := golden.VerifyOutput(mt.Terminal, goldenPath, format, trim, unwrap)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if !result.Match {
		return mcp.NewToolResultError(fmt.Sprintf("output does not match golden file %s\n%s", goldenPath, result.Diff)), nil
	}

	return mcp.NewToolResultText(result.ToJSON()), nil
}

func (s *Server) handleUpdateGolden(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	goldenPath, err := req.RequireString("golden_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	format := req.GetString("format", "plain")
	trim := req.GetBool("trim", true)
	unwrap := req.GetBool("unwrap", true)

	mt, err := s.terminals.Get(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	mt.Lock()
	defer mt.Unlock()

	if err := golden.UpdateGolden(mt.Terminal, goldenPath, format, trim, unwrap); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("golden file %q updated", goldenPath)), nil
}
