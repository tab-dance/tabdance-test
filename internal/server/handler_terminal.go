package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tab-dance/tabdance-test/internal/terminal"
)

func (s *Server) handleCreateTerminal(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cols := uint16(req.GetInt("cols", 108))
	rows := uint16(req.GetInt("rows", 24))
	scrollback := uint(req.GetInt("scrollback", 10000))
	command := req.GetString("command", "")
	workingDir := req.GetString("working_directory", "")

	var env map[string]string
	args := req.GetArguments()
	if v, ok := args["env"]; ok {
		if raw, err := json.Marshal(v); err == nil {
			if err := json.Unmarshal(raw, &env); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("parse env: %v", err)), nil
			}
		}
	}

	mt, err := s.terminals.Create(terminal.CreateConfig{
		Cols:             cols,
		Rows:             rows,
		Scrollback:       scrollback,
		Command:          command,
		Env:              env,
		WorkingDirectory: workingDir,
	})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := map[string]any{
		"terminal_id": mt.ID,
		"cols":        mt.Cols,
		"rows":        mt.Rows,
	}
	data, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handleListTerminals(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	infos := s.terminals.List()
	data, _ := json.Marshal(map[string]any{"terminals": infos})
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handleCloseTerminal(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := s.terminals.Delete(id); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("terminal %q closed", id)), nil
}

func (s *Server) handleResetTerminal(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	mt, err := s.terminals.Get(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	mt.Lock()
	defer mt.Unlock()
	mt.Terminal.Reset()
	return mcp.NewToolResultText("terminal reset"), nil
}

func (s *Server) handleResize(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	cols, err := req.RequireInt("cols")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	rows, err := req.RequireInt("rows")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if cols <= 0 || rows <= 0 {
		return mcp.NewToolResultError("cols and rows must be positive"), nil
	}

	mt, err := s.terminals.Get(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := mt.Resize(uint16(cols), uint16(rows)); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("resized to %dx%d", cols, rows)), nil
}
