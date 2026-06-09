package server

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) registerTools() {
	s.mcp.AddTool(mcp.NewTool("create_terminal",
		mcp.WithDescription("Create a new headless terminal, optionally running a child process"),
		mcp.WithNumber("cols", mcp.Description("Terminal width in columns (default: 108)")),
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
		mcp.WithDescription("Get the current terminal view as plain text, VT, HTML, or image (svg, png via freeze)"),
		mcp.WithString("terminal_id", mcp.Required(), mcp.Description("Terminal ID")),
		mcp.WithString("format", mcp.Description("Output format: plain, vt, html, svg, png (default: plain)")),
		mcp.WithBoolean("trim", mcp.Description("Trim trailing whitespace (default: true)")),
		mcp.WithBoolean("unwrap", mcp.Description("Unwrap soft-wrapped lines (default: true)")),
		mcp.WithString("output_path", mcp.Description("Output file path for image formats (svg, png). Defaults to temp file.")),
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
