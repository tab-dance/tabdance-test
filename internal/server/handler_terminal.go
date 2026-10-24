package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tab-dance/tabdance-test/internal/terminal"
)

func handleParamError(err error) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError(err.Error()), nil
}

func (s *Server) handleCreateTerminal(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfg := terminal.CreateConfig{
		Cols:             uint16(req.GetInt("cols", 108)),
		Rows:             uint16(req.GetInt("rows", 24)),
		Scrollback:       uint(req.GetInt("scrollback", 10000)),
		Command:          req.GetString("command", ""),
		WorkingDirectory: req.GetString("working_directory", ""),
	}

	if v, ok := req.GetArguments()["env"]; ok {
		var env map[string]string
		raw, _ := json.Marshal(v)
		if err := json.Unmarshal(raw, &env); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("parse env: %v", err)), nil
		}
		cfg.Env = env
	}

	created, err := s.svc.Exec(cfg, "", 0)
	if err != nil {
		return handleParamError(err)
	}
	data, _ := json.Marshal(map[string]any{
		"terminal_id": created.TerminalID,
		"cols":        created.Cols,
		"rows":        created.Rows,
	})
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handleListTerminals(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, _ := json.Marshal(map[string]any{"terminals": s.svc.List()})
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handleCloseTerminal(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return handleParamError(err)
	}
	if err := s.svc.Delete(id); err != nil {
		return handleParamError(err)
	}
	return mcp.NewToolResultText(fmt.Sprintf("terminal %q closed", id)), nil
}

func (s *Server) handleResetTerminal(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return handleParamError(err)
	}

	mt, err := s.terminals.Get(id)
	if err != nil {
		return handleParamError(err)
	}
	mt.Lock()
	defer mt.Unlock()
	mt.Terminal.Reset()
	return mcp.NewToolResultText(fmt.Sprintf("terminal %q reset", id)), nil
}

func (s *Server) handleResize(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return handleParamError(err)
	}
	cols := req.GetInt("cols", 0)
	rows := req.GetInt("rows", 0)
	if err := s.svc.Resize(id, cols, rows); err != nil {
		return handleParamError(err)
	}
	return mcp.NewToolResultText(fmt.Sprintf("terminal resized to %dx%d", cols, rows)), nil
}
