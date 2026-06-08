package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tab-dance/tabdance-test/internal/input"
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

	s.mcp.AddTool(mcp.NewTool("send_keys",
		mcp.WithDescription("Send key events to a terminal. Keys are encoded using KeyEncoder synced to terminal state."),
		mcp.WithString("terminal_id", mcp.Required(), mcp.Description("Terminal ID")),
		mcp.WithArray("keys", mcp.Required(), mcp.Description("Array of key inputs: {key, mods, text}")),
	), s.handleSendKeys)

	s.mcp.AddTool(mcp.NewTool("send_text",
		mcp.WithDescription("Send raw text to the terminal's PTY stdin (not VT-encoded)"),
		mcp.WithString("terminal_id", mcp.Required(), mcp.Description("Terminal ID")),
		mcp.WithString("text", mcp.Required(), mcp.Description("Text to send")),
	), s.handleSendText)

	s.mcp.AddTool(mcp.NewTool("send_mouse",
		mcp.WithDescription("Send mouse events to a terminal"),
		mcp.WithString("terminal_id", mcp.Required(), mcp.Description("Terminal ID")),
		mcp.WithString("action", mcp.Required(), mcp.Description("Mouse action: press, release, move")),
		mcp.WithString("button", mcp.Required(), mcp.Description("Mouse button: left, middle, right")),
		mcp.WithNumber("x", mcp.Required(), mcp.Description("X position")),
		mcp.WithNumber("y", mcp.Required(), mcp.Description("Y position")),
		mcp.WithString("mods", mcp.Description("Modifiers: ctrl, alt, shift, super (combinable with +)")),
	), s.handleSendMouse)

	s.mcp.AddTool(mcp.NewTool("send_focus",
		mcp.WithDescription("Send focus event to a terminal"),
		mcp.WithString("terminal_id", mcp.Required(), mcp.Description("Terminal ID")),
		mcp.WithString("event", mcp.Required(), mcp.Description("Focus event: gained or lost")),
	), s.handleSendFocus)

	s.mcp.AddTool(mcp.NewTool("send_paste",
		mcp.WithDescription("Send paste event (bracketed-paste-aware) to a terminal"),
		mcp.WithString("terminal_id", mcp.Required(), mcp.Description("Terminal ID")),
		mcp.WithString("text", mcp.Required(), mcp.Description("Text to paste")),
	), s.handleSendPaste)
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

func (s *Server) handleSendKeys(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	mt, err := s.terminals.Get(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var keyInputs []input.KeyInput
	if err := req.BindArguments(&struct{ Keys []input.KeyInput `json:"keys"` }{}); err != nil {
		args := req.GetArguments()
		if raw, ok := args["keys"]; ok {
			data, _ := json.Marshal(raw)
			if err := json.Unmarshal(data, &keyInputs); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("parse keys: %v", err)), nil
			}
		}
	}

	mt.Lock()
	defer mt.Unlock()

	if err := input.SendKeys(mt.KeyEncoder, mt, mt.Terminal, keyInputs); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText("keys sent"), nil
}

func (s *Server) handleSendText(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	text, err := req.RequireString("text")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	mt, err := s.terminals.Get(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	mt.Lock()
	defer mt.Unlock()

	if _, err := mt.WritePty([]byte(text)); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText("text sent"), nil
}

func (s *Server) handleSendMouse(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	action, err := req.RequireString("action")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	button, err := req.RequireString("button")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	x := float32(req.GetFloat("x", 0))
	y := float32(req.GetFloat("y", 0))
	mods := req.GetString("mods", "")

	mt, err := s.terminals.Get(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	mt.Lock()
	defer mt.Unlock()

	if err := input.SendMouse(mt.MouseEncoder, mt, input.MouseInput{
		Action: action,
		Button: button,
		X:      x,
		Y:      y,
		Mods:   mods,
	}); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText("mouse event sent"), nil
}

func (s *Server) handleSendFocus(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	event, err := req.RequireString("event")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	mt, err := s.terminals.Get(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	mt.Lock()
	defer mt.Unlock()

	if err := input.SendFocus(mt, event); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText("focus event sent"), nil
}

func (s *Server) handleSendPaste(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	text, err := req.RequireString("text")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	mt, err := s.terminals.Get(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	mt.Lock()
	defer mt.Unlock()

	if err := input.SendPaste(mt, mt.Terminal, text); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText("paste sent"), nil
}
