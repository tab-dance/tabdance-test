# tabdance-test — Implementation Plan

MCP server for headless terminal testing using go-libghostty. An AI agent can create terminals, send keys/mouse, inspect state, record sessions, play VHS tapes, and verify output against golden files.

## Architecture

```
AI Agent
  → MCP protocol (stdio)
    → tabdance-test MCP server
      → go-libghostty (headless VT terminal)
        → libghostty-vt (C VT emulation library)
      → PTY (child process running TUI app)
      → AsciicastRecorder (.cast files)
      → TapeExecutor (VHS .tape playback)
      → Golden file comparator
```

## Project Structure

```
tabdance-test/
├── cmd/
│   └── tabdance-test/
│       └── main.go              # Entrypoint: start MCP server
├── internal/
│   ├── server/
│   │   └── server.go           # MCP server setup, tool registration
│   ├── terminal/
│   │   ├── manager.go          # TerminalManager: create/list/close terminals
│   │   ├── managed.go          # ManagedTerminal: single terminal + PTY lifecycle
│   │   └── pty.go              # PTY bridge goroutines (child↔terminal)
│   ├── input/
│   │   ├── keys.go             # Key name parsing, send_keys implementation
│   │   ├── mouse.go            # Mouse event parsing, send_mouse implementation
│   │   ├── focus.go            # Focus event encoding
│   │   └── paste.go            # Bracketed-paste-aware paste encoding
│   ├── inspect/
│   │   ├── view.go             # get_view: Formatter-based screen extraction
│   │   ├── cell.go             # get_cell: GridRef-based cell inspection
│   │   ├── cursor.go           # get_cursor: cursor position/style
│   │   └── size.go             # get_size: terminal dimensions
│   ├── recording/
│   │   └── asciicast.go        # AsciicastRecorder: .cast file writing
│   ├── tape/
│   │   ├── executor.go         # TapeExecutor: VHS tape → go-libghostty ops
│   │   └── command.go          # Tape command translation table
│   ├── golden/
│   │   └── verify.go           # Golden file comparison + diff generation
│   └── wait/
│       └── wait.go             # wait_for: poll screen content for regex match
├── go.mod
├── go.sum
├── plan.md
└── skills/                      # (existing) skill references
```

## Implementation Steps

### Step 1: Project Scaffolding

- Initialize Go module: `github.com/tab-dance/tabdance-test`
- Add dependencies: `go.mitchellh.com/libghostty`, `github.com/charmbracelet/vhs/lexer`, `github.com/charmbracelet/vhs/parser`, `github.com/charmbracelet/vhs/token`, `github.com/mark3labs/mcp-go`, `github.com/creack/pty`, `github.com/pmezard/go-difflib`
- Create `cmd/tabdance-test/main.go` with MCP server bootstrap (stdio transport)
- Verify: `go build ./cmd/tabdance-test`

### Step 2: Terminal Manager + ManagedTerminal

- `internal/terminal/managed.go` — `ManagedTerminal` struct holding:
  - `*libghostty.Terminal`, `*libghostty.KeyEncoder`, `*libghostty.MouseEncoder`
  - `*exec.Cmd`, `*os.File` (ptmx), `*bytes.Buffer` (PTY output)
  - `*AsciicastRecorder`, `sync.Mutex`
  - ID, command, env, working directory
- `internal/terminal/managed.go` — `NewManagedTerminal()`:
  - Create `libghostty.NewTerminal` with `WithSize`, `WithMaxScrollback`, `WithWritePty`
  - Create `KeyEncoder` and `MouseEncoder`, sync with `SetOptFromTerminal`
  - If command provided: spawn PTY child process, start bridge goroutines
- `internal/terminal/managed.go` — `Close()`: stop recording, close encoders, kill child, close PTY, close terminal
- `internal/terminal/manager.go` — `TerminalManager` with `sync.Mutex`, `map[string]*ManagedTerminal`
  - `Create()`, `Get()`, `List()`, `Delete()`
  - Auto-generate terminal IDs
- `internal/terminal/pty.go` — Bridge goroutines:
  - Child stdout → `term.VTWrite()` + optional `recorder.WriteOutput()`
  - PTY output buffer → child stdin
- Verify: unit test creating/closing a terminal

### Step 3: MCP Tool Registration + Terminal Management Tools

- `internal/server/server.go` — Create MCP server, register all tools
- Implement MCP tools:
  - `create_terminal` — cols, rows, scrollback, command, env, working_directory → terminal_id
  - `list_terminals` — list active terminals
  - `close_terminal` — close and destroy
  - `reset_terminal` — `term.Reset()`
  - `resize` — `term.Resize()` + PTY resize + encoder re-sync
- Verify: start server, create/list/close terminals via MCP

### Step 4: Input Injection

- `internal/input/keys.go` — Key name → `libghostty.Key` mapping (W3C names)
- `internal/input/keys.go` — Modifier string → `libghostty.Mods` parsing
- `internal/input/keys.go` — `SendKeys()`: iterate key inputs, encode via KeyEncoder, write to PTY
  - Handle `text` field (raw text to PTY)
  - Send press + release for Kitty keyboard protocol
- `internal/input/mouse.go` — `SendMouse()`: parse action/button/mods, encode via MouseEncoder
- `internal/input/focus.go` — `SendFocus()`: encode focus gained/lost
- `internal/input/paste.go` — `SendPaste()`: bracketed-paste-aware encoding
- MCP tools: `send_keys`, `send_text`, `send_mouse`, `send_focus`, `send_paste`
- Verify: create terminal with bash, send keys, read view

### Step 5: Terminal Inspection

- `internal/inspect/view.go` — `GetView()`: Formatter with plain/vt/html, trim, unwrap
- `internal/inspect/cell.go` — `GetCell()`: GridRef for specific cell (codepoint, wide, bold, italic, fg, bg, hyperlink)
- `internal/inspect/cursor.go` — `GetCursor()`: x, y, visible, pending_wrap, style
- `internal/inspect/size.go` — `GetSize()`: cols, rows, width_px, height_px
- MCP tools: `get_view`, `get_cell`, `get_cursor`, `get_size`
- Verify: create terminal, write VT data, inspect content

### Step 6: Asciinema Recording

- `internal/recording/asciicast.go` — `AsciicastRecorder`:
  - Write header (version 2, width, height, env, idle_time_limit)
  - `WriteOutput(data)`, `WriteInput(data)`, `WriteMarker(label)`, `WriteResize(cols, rows)`
  - `Close()` to flush and close file
- Integrate with `ManagedTerminal`:
  - `StartRecording(path, idleLimit)` / `StopRecording()`
  - PTY bridge goroutines call `recorder.WriteOutput()` when recorder is active
- MCP tools: `start_recording`, `stop_recording`, `add_marker`
- Verify: record a session, play back with `asciinema play`

### Step 7: VHS Tape Playback

- `internal/tape/executor.go` — `TapeExecutor`:
  - Parse .tape file using `vhs/lexer` + `vhs/parser`
  - Execute commands sequentially with timeout
- `internal/tape/command.go` — Command translation:
  - `Type` → KeyEncoder per character with typing speed delay
  - `Enter/Backspace/Tab/Space` → encode corresponding key
  - `Ctrl+X/Alt+X/Shift+X` → encode with modifiers
  - `Arrow keys/PageUp/PageDown` → encode navigation keys
  - `Escape/Delete/Insert` → encode special keys
  - `Sleep` → `time.Sleep()`
  - `Wait /pattern/` → poll `GetView()` with regex
  - `Set Width/Height` → resize
  - `Set TypingSpeed` → update per-character delay
  - `Set FontSize/Shell` → no-op (headless)
  - `Screenshot` → RenderState-based capture (placeholder)
  - `Hide/Show` → pause/resume recording
  - `ScrollUp/ScrollDown` → `ScrollViewportDelta()`
  - `Source` → parse and execute included tape
  - `Env` → set env var (pre-start only)
- MCP tool: `play_tape`
- Verify: create terminal, play a .tape file, check result

### Step 8: Golden File Verification

- `internal/golden/verify.go` — `VerifyOutput()`:
  - Get current terminal content in specified format
  - Read golden file (create if not exists + update flag)
  - Compare content
  - Generate unified diff on mismatch using `go-difflib`
  - Return match/mismatch + diff
- MCP tools: `verify_output`, `update_golden`
- Verify: create terminal, write content, verify against golden file

### Step 9: Wait/Poll

- `internal/wait/wait.go` — `WaitFor()`:
  - Poll `GetView()` at interval (50ms)
  - Match regex against screen (scope=screen) or current line (scope=line)
  - Timeout with error
  - Return found/not-found + elapsed time
- MCP tool: `wait_for`
- Verify: create terminal with slow command, wait for pattern

### Step 10: Integration Testing + Polish

- End-to-end test: create terminal → run command → send keys → inspect → verify golden → record → close
- Test with tabdance: `tabdance select --out bash:/tmp/td-out`
- Test VHS tape playback against real .tape file
- Error handling: invalid terminal IDs, closed terminals, PTY failures
- Graceful shutdown: close all terminals on server exit
- Verify: full integration test passes

## Dependencies

| Package | Purpose |
|---------|---------|
| `go.mitchellh.com/libghostty` | Headless VT terminal, KeyEncoder, MouseEncoder, Formatter, GridRef, RenderState |
| `github.com/mark3labs/mcp-go` | MCP server framework (tool registration, stdio transport) |
| `github.com/creack/pty` | PTY creation for child process spawning |
| `github.com/charmbracelet/vhs/lexer` | VHS tape lexer |
| `github.com/charmbracelet/vhs/parser` | VHS tape parser |
| `github.com/charmbracelet/vhs/token` | VHS token types |
| `github.com/pmezard/go-difflib` | Unified diff generation for golden file comparison |

## MCP Tool Summary

| Tool | Input | Output |
|------|-------|--------|
| `create_terminal` | cols, rows, scrollback, command, env, working_directory | terminal_id, cols, rows |
| `list_terminals` | — | terminals[] |
| `close_terminal` | terminal_id | — |
| `reset_terminal` | terminal_id | — |
| `resize` | terminal_id, cols, rows | — |
| `send_keys` | terminal_id, keys[] | — |
| `send_text` | terminal_id, text | — |
| `send_mouse` | terminal_id, action, button, x, y, mods | — |
| `send_focus` | terminal_id, event | — |
| `send_paste` | terminal_id, text | — |
| `get_view` | terminal_id, format, trim, unwrap | content, cursor_x, cursor_y, cursor_visible, cursor_style, active_screen |
| `get_cell` | terminal_id, x, y | codepoint, text, wide, bold, italic, fg, bg, hyperlink |
| `get_cursor` | terminal_id | x, y, visible, style, pending_wrap |
| `get_size` | terminal_id | cols, rows, width_px, height_px |
| `start_recording` | terminal_id, output_path, idle_time_limit | — |
| `stop_recording` | terminal_id | — |
| `add_marker` | terminal_id, label | — |
| `play_tape` | terminal_id, tape_path, timeout | — |
| `verify_output` | terminal_id, golden_path, format, trim, unwrap, update | match, diff |
| `update_golden` | terminal_id, golden_path, format, trim, unwrap | — |
| `wait_for` | terminal_id, pattern, scope, timeout | found, elapsed_ms |

## Key Design Decisions

1. **Per-terminal mutex**: Each `ManagedTerminal` has its own `sync.Mutex` — allows concurrent operations on different terminals while serializing access to each one
2. **PTY bridge goroutines**: Separate goroutines for child→terminal and terminal→child data flow, with recorder integration in the child→terminal path
3. **VHS parser reuse**: Use `charmbracelet/vhs/lexer` + `vhs/parser` directly rather than reimplementing tape parsing
4. **Encoder sync after resize**: Always call `SetOptFromTerminal` after resize to keep KeyEncoder/MouseEncoder in sync with terminal mode state
5. **Kitty keyboard protocol**: Send both press and release events when Kitty keyboard flags are active
6. **Golden file update mode**: Separate `update_golden` tool (not a flag on `verify_output`) to make accidental updates harder
