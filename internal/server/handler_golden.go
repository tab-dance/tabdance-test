package server

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handleVerifyOutput(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return handleParamError(err)
	}
	goldenPath, err := req.RequireString("golden_path")
	if err != nil {
		return handleParamError(err)
	}
	format := req.GetString("format", "plain")
	trim := req.GetBool("trim", true)
	unwrap := req.GetBool("unwrap", true)

	result, err := s.svc.Verify(id, goldenPath, format, trim, unwrap)
	if err != nil {
		return handleParamError(err)
	}

	if !result.Match {
		return mcp.NewToolResultError(fmt.Sprintf("output does not match golden file %s\n%s", goldenPath, result.Diff)), nil
	}
	return mcp.NewToolResultText(result.ToJSON()), nil
}

func (s *Server) handleUpdateGolden(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return handleParamError(err)
	}
	goldenPath, err := req.RequireString("golden_path")
	if err != nil {
		return handleParamError(err)
	}
	format := req.GetString("format", "plain")
	trim := req.GetBool("trim", true)
	unwrap := req.GetBool("unwrap", true)

	if err := s.svc.UpdateGolden(id, goldenPath, format, trim, unwrap); err != nil {
		return handleParamError(err)
	}
	return mcp.NewToolResultText(fmt.Sprintf("golden file %q updated", goldenPath)), nil
}
