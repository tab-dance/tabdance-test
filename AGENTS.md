# AGENTS.md

## Project Overview

tabdance-test is an MCP server for headless terminal testing using [go-libghostty](https://go.mitchellh.com/libghostty). An AI agent can create terminals, send keys/mouse, inspect state, record sessions, play VHS tapes, and verify output against golden files. Designed for testing Bubble Tea applications (e.g., tabdance) via the MCP protocol.

## Essential Commands

```sh
make build-static           # Build with static libghostty-vt linking (default)
make build-dynamic         # Build with dynamic linking (needs LD_LIBRARY_PATH)
make test                  # Run tests (needs PKG_CONFIG_PATH + LD_LIBRARY_PATH)
make clean                 # Remove binary
make run                   # Build dynamic + run
```

Build requires `PKG_CONFIG_PATH` pointing to ghostty's `zig-out/share/pkgconfig` (set via `GHOSTTY_DIR` in Makefile, defaults to `~/Documents/development/github/ghostty-org/ghostty`).

## Architecture & Data Flow

```
AI Agent
  → MCP protocol (stdio)
    → tabdance-test MCP server (mark3labs/mcp-go)
      → go-libghostty (headless VT terminal)
        → libghostty-vt (C VT emulation library, built from ghostty source)
      → PTY (creack/pty, child process running TUI app)
      → AsciicastRecorder (.cast files)
      → TapeExecutor (VHS .tape playback via charmbracelet/vhs parser)
      → Golden file comparator (go-difflib)
      → WaitFor (regex polling)
```

### Key flow: create terminal with command

1. `libghostty.NewTerminal()` with `WithSize`, `WithMaxScrollback`, `WithWritePty`
2. Create `KeyEncoder` + `MouseEncoder`, sync with `SetOptFromTerminal`
3. If command provided: spawn PTY child process, start bridge goroutines
4. Bridge: child stdout → `term.VTWrite()` + optional `recorder.WriteOutput()`
5. Bridge: `WritePty` callback → ptyOut buffer → `ptmx.Write()` → child stdin

### Key flow: send keys

1. `libghostty.ParseKey(key)` + `libghostty.ParseMods(mods)` — parse W3C key names
2. `libghostty.NewKeyEvent()` → set action/key/mods → `KeyEncoder.Encode()`
3. Write encoded bytes to PTY master
4. If Kitty keyboard active: also send release event

## Code Organization

```
cmd/tabdance-test/
  main.go                    # Entrypoint: server.New().Run()

internal/
  server/
    server.go                # MCP server setup, tool registration, all 20 tool handlers
    server_test.go            # Integration tests (create, list, view, cursor, size, resize, invalid ID, golden)
  terminal/
    manager.go               # TerminalManager: Create/Get/List/Delete/CloseAll, auto-increment IDs
    managed.go                # ManagedTerminal: terminal + PTY lifecycle, encoders, recorder, mutex
  input/
    keys.go                  # SendKeys: ParseKey/ParseMods, Kitty keyboard press+release
    mouse.go                  # SendMouse: ParseMouseButton, scroll_up/scroll_down via button 4/5
    focus.go                  # SendFocus: ParseFocusEvent + FocusEncode
    paste.go                  # SendPaste: PasteEncode with bracketed-paste mode
  inspect/
    view.go                   # GetView: Formatter (plain/vt/html, trim, unwrap) + cursor + active screen
    cell.go                   # GetCell: GridRef (codepoint, graphemes, wide, bold, italic, fg/bg, hyperlink)
    cursor.go                 # GetCursor: x, y, visible, pending_wrap, style
    size.go                   # GetSize: cols, rows, width_px, height_px
  recording/
    asciicast.go              # AsciicastRecorder: asciicast v2 .cast files (output/input/marker/resize)
  tape/
    executor.go               # TapeExecutor: VHS tape → go-libghostty ops, Terminal interface
  golden/
    verify.go                 # VerifyOutput/UpdateGolden: go-difflib unified diff
  wait/
    wait.go                   # WaitFor: regex polling at 50ms, screen/line scope

skills/                       # Crush skill references (ghostty, asciinema, vhs)
```

## MCP Tool Summary

20 tools registered in `server.go:registerTools()`:

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
| `verify_output` | terminal_id, golden_path, format, trim, unwrap | match, diff |
| `update_golden` | terminal_id, golden_path, format, trim, unwrap | — |
| `wait_for` | terminal_id, pattern, scope, timeout, format, trim, unwrap | found, elapsed_ms |

## Conventions & Patterns

- **Per-terminal mutex**: Each `ManagedTerminal` has its own `sync.Mutex`, exposed via `Lock()`/`Unlock()`. MCP handlers lock for the duration of their operation.
- **io.Writer on ManagedTerminal**: `Write()` delegates to `WritePty()` so input package can use `io.Writer` interface.
- **Encoder sync after resize**: Always call `SetOptFromTerminal` after resize to keep KeyEncoder/MouseEncoder in sync with terminal mode state.
- **Kitty keyboard protocol**: Send both press and release events when `KittyKeyboardFlags() != KittyKeyDisabled`.
- **VHS parser reuse**: Use `charmbracelet/vhs/lexer` + `vhs/parser` directly. API: `parser.New(lexer)` + `p.Parse()` (check `p.Errors()` separately).
- **TapeExecutor Terminal interface**: Decoupled from `ManagedTerminal` via interface with `Lock/Unlock/WritePty/GetTerminal/GetKeyEncoder/GetMouseEncoder/Resize/AddMarker`.
- **Golden file update mode**: Separate `update_golden` tool (not a flag on `verify_output`) to make accidental updates harder.
- **Scroll mouse actions**: `scroll_up`/`scroll_down` translated to button 4/5 press+release sequences.
- **All go-libghostty getters return `(value, error)`**: e.g., `CursorX() (uint16, error)`, `Cols() (uint16, error)`.
- **mcp-go CallToolRequest**: Use `req.GetString()`, `req.GetInt()`, `req.GetBool()`, `req.RequireString()`, `req.RequireInt()` — NOT `req.Params.Arguments["key"]`.

## Gotchas

- **libghostty-vt must be built from source**: System-installed version is too old. Must build from ghostty source at the commit pinned by go-libghostty (`cb36966a752982014827a9cabcf630ec3788b3d9`).
- **Zig version**: System Zig 0.16.0 is incompatible. Use Zig 0.15.2 (downloaded to `/tmp/zig-x86_64-linux-0.15.2/`).
- **Build command**: `make build-static` (sets PKG_CONFIG_PATH automatically). Don't use bare `go build`.
- **`KeyUnknown` doesn't exist**: Use `KeyUnidentified` (value 0) as the zero/unknown sentinel for `libghostty.Key`.
- **`MousePosition` uses `float32`**: Not `uint32` or `int`.
- **`Graphemes()` returns `[]uint32`**: Not a string — must convert codepoints to runes.
- **`CellWideWide`**: Not `CellWideYes`. Other constants: `CellWideNarrow`, `CellWideSpacerHead`, `CellWideSpacerTail`.
- **`StyleColor` is a struct**: With `Tag` (None/Palette/RGB), `Palette uint8`, `RGB ColorRGB`.
- **VHS parser API**: `parser.New(l)` takes lexer as argument; `p.Parse()` returns `[]Command` with no error return — check `p.Errors()` separately.
- **`go mod tidy` warnings**: The `go.mod` has indirect deps that gopls complains about — these come from mcp-go's transitive dependencies and are normal.
- **PTY bridge goroutines**: Exit when `done` channel is closed (in `Close()`). The `WritePty` callback runs synchronously during `VTWrite`.
- **No `MouseActionScrollUp/ScrollDown`**: go-libghostty only has Press/Release/Motion. Scroll is implemented via button 4/5 press+release.

## Dependencies

| Package | Purpose |
|---------|---------|
| `go.mitchellh.com/libghostty` | Headless VT terminal, KeyEncoder, MouseEncoder, Formatter, GridRef |
| `github.com/mark3labs/mcp-go` | MCP server framework (tool registration, stdio transport) |
| `github.com/creack/pty` | PTY creation for child process spawning |
| `github.com/charmbracelet/vhs` | VHS tape parser (lexer, parser, token sub-packages) |
| `github.com/pmezard/go-difflib` | Unified diff generation for golden file comparison |

## Testing

```sh
make test  # Runs go test ./... with PKG_CONFIG_PATH and LD_LIBRARY_PATH
```

Integration tests in `internal/server/server_test.go` cover: create/close, list, get_view, get_cursor, get_size, resize, invalid terminal ID, and golden file verification for nonexistent files.

## Cross-Project Skills

- **ghostty** skill — go-libghostty API details (Terminal, KeyEncoder, MouseEncoder, Formatter, GridRef, etc.)
- **asciinema** skill — asciicast v2 format specification
- **vhs** skill — VHS tape DSL and parsing
- **bubbletea** skill — Bubble Tea framework internals
- **tabdance** skill — tabdance application internals
