package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tab-dance/tabdance-test/internal/terminal"
)

type Server struct {
	mcp      *server.MCPServer
	terminals *terminal.TerminalManager
}

func New() *Server {
	s := &Server{
		terminals: terminal.NewTerminalManager(),
	}

	s.mcp = server.NewMCPServer(
		"tabdance-test",
		"0.1.0",
		server.WithToolCapabilities(false),
	)

	s.registerTools()

	return s
}

func (s *Server) Run() error {
	return server.ServeStdio(s.mcp)
}

func (s *Server) registerTools() {
	s.mcp.AddTool(mcp.NewTool("create_terminal",
		mcp.WithDescription("Create a new headless terminal, optionally running a child process"),
		mcp.WithNumber("cols", mcp.Description("Terminal width in columns (default: 80)")),
		mcp.WithNumber("rows", mcp.Description("Terminal height in rows (default: 24)")),
		mcp.WithNumber("scrollback", mcp.Description("Max scrollback lines (default: 10000)")),
		mcp.WithString("command", mcp.Description("Command to run in the terminal")),
		mcp.WithObject("env", mcp.Description("Environment variables (key-value pairs)")),
		mcp.WithString("working_directory", mcp.Description("Working directory for the child process")),
	), s.handleCreateTerminal)

	s.mcp.AddTool(mcp.NewTool("list_terminals",
		mcp.WithDescription("List all active terminals"),
	), s.handleListTerminals)

	s.mcp.AddTool(mcp.NewTool("close_terminal",
		mcp.WithDescription("Close and destroy a terminal"),
		mcp.WithString("terminal_id", mcp.Required(), mcp.Description("Terminal ID to close")),
	), s.handleCloseTerminal)

	s.mcp.AddTool(mcp.NewTool("reset_terminal",
		mcp.WithDescription("Full terminal reset (RIS)"),
		mcp.WithString("terminal_id", mcp.Required(), mcp.Description("Terminal ID to reset")),
	), s.handleResetTerminal)

	s.mcp.AddTool(mcp.NewTool("resize",
		mcp.WithDescription("Resize the terminal"),
		mcp.WithString("terminal_id", mcp.Required(), mcp.Description("Terminal ID to resize")),
		mcp.WithNumber("cols", mcp.Required(), mcp.Description("New width in columns")),
		mcp.WithNumber("rows", mcp.Required(), mcp.Description("New height in rows")),
	), s.handleResize)
}

func (s *Server) handleCreateTerminal(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cols := uint16(req.GetInt("cols", 80))
	rows := uint16(req.GetInt("rows", 24))
	scrollback := uint(req.GetInt("scrollback", 10000))
	command := req.GetString("command", "")
	workingDir := req.GetString("working_directory", "")

	var env map[string]string
	args := req.GetArguments()
	if v, ok := args["env"]; ok {
		if raw, err := json.Marshal(v); err == nil {
			json.Unmarshal(raw, &env)
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

	result := map[string]interface{}{
		"terminal_id": mt.ID,
		"cols":        mt.Cols,
		"rows":        mt.Rows,
	}
	data, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handleListTerminals(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	infos := s.terminals.List()
	data, _ := json.Marshal(map[string]interface{}{"terminals": infos})
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
