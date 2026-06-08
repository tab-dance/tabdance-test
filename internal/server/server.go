package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tab-dance/tabdance-test/internal/golden"
	"github.com/tab-dance/tabdance-test/internal/input"
	"github.com/tab-dance/tabdance-test/internal/inspect"
	"github.com/tab-dance/tabdance-test/internal/tape"
	"github.com/tab-dance/tabdance-test/internal/terminal"
	"github.com/tab-dance/tabdance-test/internal/wait"
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
	defer s.terminals.CloseAll()
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

	s.mcp.AddTool(mcp.NewTool("get_view",
		mcp.WithDescription("Get the current terminal view as plain text, VT, or HTML"),
		mcp.WithString("terminal_id", mcp.Required(), mcp.Description("Terminal ID")),
		mcp.WithString("format", mcp.Description("Output format: plain, vt, html (default: plain)")),
		mcp.WithBoolean("trim", mcp.Description("Trim trailing whitespace (default: true)")),
		mcp.WithBoolean("unwrap", mcp.Description("Unwrap soft-wrapped lines (default: true)")),
	), s.handleGetView)

	s.mcp.AddTool(mcp.NewTool("get_cell",
		mcp.WithDescription("Get a specific cell's content and style"),
		mcp.WithString("terminal_id", mcp.Required(), mcp.Description("Terminal ID")),
		mcp.WithNumber("x", mcp.Required(), mcp.Description("Column (0-indexed)")),
		mcp.WithNumber("y", mcp.Required(), mcp.Description("Row (0-indexed)")),
	), s.handleGetCell)

	s.mcp.AddTool(mcp.NewTool("get_cursor",
		mcp.WithDescription("Get cursor position and style"),
		mcp.WithString("terminal_id", mcp.Required(), mcp.Description("Terminal ID")),
	), s.handleGetCursor)

	s.mcp.AddTool(mcp.NewTool("get_size",
		mcp.WithDescription("Get terminal dimensions"),
		mcp.WithString("terminal_id", mcp.Required(), mcp.Description("Terminal ID")),
	), s.handleGetSize)

	s.mcp.AddTool(mcp.NewTool("start_recording",
		mcp.WithDescription("Start asciinema recording for a terminal"),
		mcp.WithString("terminal_id", mcp.Required(), mcp.Description("Terminal ID")),
		mcp.WithString("output_path", mcp.Required(), mcp.Description("Path to write .cast file")),
		mcp.WithNumber("idle_time_limit", mcp.Description("Idle time limit in seconds (default: 2.0)")),
	), s.handleStartRecording)

	s.mcp.AddTool(mcp.NewTool("stop_recording",
		mcp.WithDescription("Stop recording and save the .cast file"),
		mcp.WithString("terminal_id", mcp.Required(), mcp.Description("Terminal ID")),
	), s.handleStopRecording)

	s.mcp.AddTool(mcp.NewTool("add_marker",
		mcp.WithDescription("Add a marker to the current recording"),
		mcp.WithString("terminal_id", mcp.Required(), mcp.Description("Terminal ID")),
		mcp.WithString("label", mcp.Required(), mcp.Description("Marker label")),
	), s.handleAddMarker)

	s.mcp.AddTool(mcp.NewTool("play_tape",
		mcp.WithDescription("Play a VHS tape file (.tape) in a terminal, executing each command sequentially"),
		mcp.WithString("terminal_id", mcp.Required(), mcp.Description("Terminal ID")),
		mcp.WithString("tape_path", mcp.Required(), mcp.Description("Path to the .tape file")),
		mcp.WithNumber("timeout", mcp.Description("Timeout in seconds for Wait commands (default: 30)")),
	), s.handlePlayTape)

	s.mcp.AddTool(mcp.NewTool("verify_output",
		mcp.WithDescription("Verify terminal output against a golden file. Returns match status and diff on mismatch."),
		mcp.WithString("terminal_id", mcp.Required(), mcp.Description("Terminal ID")),
		mcp.WithString("golden_path", mcp.Required(), mcp.Description("Path to the golden file")),
		mcp.WithString("format", mcp.Description("Output format: plain, vt, html (default: plain)")),
		mcp.WithBoolean("trim", mcp.Description("Trim trailing whitespace (default: true)")),
		mcp.WithBoolean("unwrap", mcp.Description("Unwrap soft-wrapped lines (default: true)")),
	), s.handleVerifyOutput)

	s.mcp.AddTool(mcp.NewTool("update_golden",
		mcp.WithDescription("Update (or create) a golden file with the current terminal output"),
		mcp.WithString("terminal_id", mcp.Required(), mcp.Description("Terminal ID")),
		mcp.WithString("golden_path", mcp.Required(), mcp.Description("Path to the golden file")),
		mcp.WithString("format", mcp.Description("Output format: plain, vt, html (default: plain)")),
		mcp.WithBoolean("trim", mcp.Description("Trim trailing whitespace (default: true)")),
		mcp.WithBoolean("unwrap", mcp.Description("Unwrap soft-wrapped lines (default: true)")),
	), s.handleUpdateGolden)

	s.mcp.AddTool(mcp.NewTool("wait_for",
		mcp.WithDescription("Poll terminal output until a regex pattern matches or timeout"),
		mcp.WithString("terminal_id", mcp.Required(), mcp.Description("Terminal ID")),
		mcp.WithString("pattern", mcp.Required(), mcp.Description("Regex pattern to match")),
		mcp.WithString("scope", mcp.Description("Match scope: screen (default) or line")),
		mcp.WithNumber("timeout", mcp.Description("Timeout in seconds (default: 10)")),
		mcp.WithString("format", mcp.Description("Output format: plain, vt, html (default: plain)")),
		mcp.WithBoolean("trim", mcp.Description("Trim trailing whitespace (default: true)")),
		mcp.WithBoolean("unwrap", mcp.Description("Unwrap soft-wrapped lines (default: true)")),
	), s.handleWaitFor)
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

func (s *Server) handleGetView(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	mt, err := s.terminals.Get(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	format := req.GetString("format", "plain")
	trim := req.GetBool("trim", true)
	unwrap := req.GetBool("unwrap", true)

	mt.Lock()
	defer mt.Unlock()

	result, err := inspect.GetView(mt.Terminal, format, trim, unwrap)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(result.ToJSON()), nil
}

func (s *Server) handleGetCell(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	x, err := req.RequireInt("x")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	y, err := req.RequireInt("y")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	mt, err := s.terminals.Get(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	mt.Lock()
	defer mt.Unlock()

	result, err := inspect.GetCell(mt.Terminal, uint16(x), uint32(y))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(result.ToJSON()), nil
}

func (s *Server) handleGetCursor(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	result := inspect.GetCursor(mt.Terminal)
	return mcp.NewToolResultText(result.ToJSON()), nil
}

func (s *Server) handleGetSize(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	result := inspect.GetSize(mt.Terminal)
	return mcp.NewToolResultText(result.ToJSON()), nil
}

func (s *Server) handleStartRecording(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	path, err := req.RequireString("output_path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	idleLimit := req.GetFloat("idle_time_limit", 2.0)

	mt, err := s.terminals.Get(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := mt.StartRecording(path, idleLimit); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("recording to %s", path)), nil
}

func (s *Server) handleStopRecording(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	mt, err := s.terminals.Get(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := mt.StopRecording(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText("recording stopped"), nil
}

func (s *Server) handleAddMarker(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	label, err := req.RequireString("label")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	mt, err := s.terminals.Get(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := mt.AddMarker(label); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("marker %q added", label)), nil
}

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

	mt.Lock()
	defer mt.Unlock()

	result, err := wait.WaitFor(mt.Terminal, wait.WaitForOptions{
		Pattern:      pattern,
		Scope:        scope,
		Timeout:      timeout,
		Format:       format,
		Trim:         trim,
		Unwrap:       unwrap,
	})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(result.ToJSON()), nil
}
