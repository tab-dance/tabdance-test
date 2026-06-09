# AGENTS.md

## Project Overview

tabdance-test is a CLI tool and MCP server for headless terminal testing using [go-libghostty](https://go.mitchellh.com/libghostty). It exposes two interfaces:

1. **CLI** (Cobra + carapace completions) — for local development and CI testing
2. **MCP server** (stdio) — for AI agent–driven testing

Both interfaces provide the same capabilities: create terminals, send keys/mouse, inspect state, record sessions, play VHS tapes, and verify output against golden files.

## Essential Commands

```sh
make build-static           # Build libghostty-vt + static link (default)
make build-dynamic         # Build libghostty-vt + dynamic link
make build-lib             # Build libghostty-vt from submodule only
make test                  # Build + run tests
make clean                 # Remove binary
make run                   # Build dynamic + run
```

Ghostty is included as a git submodule at `deps/ghostty` (pinned to commit `cb36966a752982014827a9cabcf630ec3788b3d9`). The `build-lib` target compiles `libghostty-vt` from the submodule using Zig. All other build targets depend on `build-lib` automatically. Requires Zig 0.15.2 (set via `ZIG` make variable, defaults to `/tmp/zig-x86_64-linux-0.15.2/zig`).

## Architecture & Data Flow

### CLI Mode (daemon-backed)

```
CLI client (tabdance-test view, keys, etc.)
  → Unix socket (JSON-RPC)
    → Daemon process (auto-forked, holds TerminalManager)
      → go-libghostty (headless VT terminal)
        → libghostty-vt (C VT emulation library, built from ghostty source)
      → PTY (creack/pty, child process running TUI app)
      → AsciicastRecorder (.cast files)
      → TapeExecutor (VHS .tape playback)
      → Golden file comparator (go-difflib)
      → WaitFor (regex polling)
```

### CLI Mode (daemon-free, `run` command)

```
tabdance-test run tape.tape --golden test.golden
  → in-process: create terminal → play tape → verify golden → exit 0/1
```

### MCP Mode

```
AI Agent
  → MCP protocol (stdio)
    → tabdance-test mcp
      → (same internal packages as daemon)
```

### Key flow: daemon auto-start

1. CLI command invoked (e.g. `tabdance-test view term_1`)
2. Client checks if daemon is running (PID file + socket connect)
3. If not: fork daemon process (`tabdance-test --daemon`), wait for socket
4. Connect, send JSON-RPC request, receive response, disconnect
5. Daemon idles out after 30s with no active terminals or clients

### Key flow: create terminal with command

1. `libghostty.NewTerminal()` with `WithSize`, `WithMaxScrollback`, `WithWritePty`
2. Create `KeyEncoder` + `MouseEncoder`, sync with `SetOptFromTerminal`
3. If command provided: spawn PTY child process, start bridge goroutines
4. Bridge: child stdout → `term.VTWrite()` + optional `recorder.WriteOutput()` + broadcast to subscribers
5. Bridge: `WritePty` callback → ptyOut buffer → `ptmx.Write()` → child stdin

### Key flow: send keys

1. `libghostty.ParseKey(key)` + `libghostty.ParseMods(mods)` — parse W3C key names
2. `libghostty.NewKeyEvent()` → set action/key/mods → `KeyEncoder.Encode()`
3. Write encoded bytes to PTY master
4. If Kitty keyboard active: also send release event

## CLI Command Summary

```
tabdance-test
├── mcp                          # Start MCP server (stdio)
├── exec <command>               # Create terminal, print ID
│   --cols / --rows / --scrollback / --env / --working-directory
│   --record <path> / --detach / -d (default: true)
├── ls                           # List active terminals (--json)
├── inspect <id>                 # Terminal details (--json)
├── attach <id>                  # Stream PTY output (Ctrl+C to detach)
├── kill <id>                    # Kill terminal
├── rm <id>                      # Remove/close terminal
├── view <id>                    # Print screen content (--format, --json)
├── cursor <id>                  # Print cursor state (--json)
├── size <id>                    # Print dimensions (--json)
├── cell <id> <x> <y>           # Inspect cell (--json)
├── send <id> <text>            # Send raw text
├── keys <id> <key...>          # Send key events
├── mouse <id> <action> <button> <x> <y>  # Send mouse event
├── paste <id> <text>           # Send paste event
├── focus <id> <gained|lost>    # Send focus event
├── resize <id> <cols> <rows>   # Resize terminal
├── tape <id> <path>            # Play .tape file (--timeout)
├── wait <id> <pattern>         # Wait for regex (--scope, --timeout)
├── verify <id> <golden>        # Verify against golden file
├── snapshot <id> <golden>      # Update/create golden file
├── record <id> <path>          # Start asciinema recording
├── stop <id>                    # Stop recording
├── marker <id> <label>          # Add recording marker
└── run <tape>                   # Daemon-free: play tape, verify, exit 0/1
    --golden / --update / --cols / --rows / --timeout / --record / --env
```

## Code Organization

```
cmd/tabdance-test/
  main.go                    # Entrypoint: --daemon flag
  cmd/
    root.go                  # Root command var + carapace integration
    mcp.go                   # `mcp` subcommand (init-style self-registration)
    exec.go                  # `exec` subcommand
    ls.go                    # `ls` subcommand
    inspect.go               # `inspect` subcommand
    attach.go                # `attach` subcommand
    kill.go                  # `kill` subcommand
    rm.go                    # `rm` subcommand
    view.go                  # `view` subcommand
    cursor.go                # `cursor` subcommand
    size.go                  # `size` subcommand
    cell.go                  # `cell` subcommand
    send.go                  # `send` subcommand
    keys.go                  # `keys` subcommand
    mouse.go                 # `mouse` subcommand
    paste.go                 # `paste` subcommand
    focus.go                 # `focus` subcommand
    resize.go                # `resize` subcommand
    tape.go                  # `tape` subcommand
    wait.go                  # `wait` subcommand
    verify.go                # `verify` subcommand
    snapshot.go              # `snapshot` subcommand
    record.go                # `record` subcommand
    stop.go                  # `stop` subcommand
    marker.go                # `marker` subcommand
    run.go                   # `run` subcommand (daemon-free, in-process runner)
    client.go                # newClient() and parseKeyValue() helpers

pkg/
  actions/
    completion.go             # Public carapace completion actions (ActionTerminalIDs, ActionKeyNames)

internal/
  daemon/
    protocol.go              # JSON-RPC protocol types (Request, Response)
    daemon.go                # Daemon: socket listener, accept loop, idle timeout, auto-start
    handler.go                # Request handler: dispatches to TerminalManager ops
  client/
    client.go                # Daemon client: connect, send request, receive response
    daemon.go                 # Auto-start logic: EnsureDaemon, IsDaemonRunning
    attach.go                 # Streaming attach client
  server/
    server.go                # MCP server setup, tool registration, all 20 tool handlers
    server_test.go            # Integration tests
  terminal/
    manager.go               # TerminalManager: Create/Get/List/Delete/CloseAll, auto-increment IDs
    managed.go                # ManagedTerminal: terminal + PTY lifecycle, encoders, recorder, mutex, subscribers
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

|| Tool | Input | Output |
||------|-------|--------|
|| `create_terminal` | cols, rows, scrollback, command, env, working_directory | terminal_id, cols, rows |
|| `list_terminals` | — | terminals[] |
|| `close_terminal` | terminal_id | — |
|| `reset_terminal` | terminal_id | — |
|| `resize` | terminal_id, cols, rows | — |
|| `send_keys` | terminal_id, keys[] | — |
|| `send_text` | terminal_id, text | — |
|| `send_mouse` | terminal_id, action, button, x, y, mods | — |
|| `send_focus` | terminal_id, event | — |
|| `send_paste` | terminal_id, text | — |
|| `get_view` | terminal_id, format, trim, unwrap | content, cursor_x, cursor_y, cursor_visible, cursor_style, active_screen |
|| `get_cell` | terminal_id, x, y | codepoint, text, wide, bold, italic, fg, bg, hyperlink |
|| `get_cursor` | terminal_id | x, y, visible, style, pending_wrap |
|| `get_size` | terminal_id | cols, rows, width_px, height_px |
|| `start_recording` | terminal_id, output_path, idle_time_limit | — |
|| `stop_recording` | terminal_id | — |
|| `add_marker` | terminal_id, label | — |
|| `play_tape` | terminal_id, tape_path, timeout | — |
|| `verify_output` | terminal_id, golden_path, format, trim, unwrap | match, diff |
|| `update_golden` | terminal_id, golden_path, format, trim, unwrap | — |
|| `wait_for` | terminal_id, pattern, scope, timeout, format, trim, unwrap | found, elapsed_ms |

## Conventions & Patterns

- **Per-terminal mutex**: Each `ManagedTerminal` has its own `sync.Mutex`, exposed via `Lock()`/`Unlock()`. MCP handlers and daemon handlers lock for the duration of their operation.
- **io.Writer on ManagedTerminal**: `Write()` delegates to `WritePty()` so input package can use `io.Writer` interface.
- **Encoder sync after resize**: Always call `SetOptFromTerminal` after resize to keep KeyEncoder/MouseEncoder in sync with terminal mode state.
- **Kitty keyboard protocol**: Send both press and release events when `KittyKeyboardFlags() != KittyKeyDisabled`.
- **VHS parser reuse**: Use `charmbracelet/vhs/lexer` + `vhs/parser` directly. API: `parser.New(lexer)` + `p.Parse()` (check `p.Errors()` separately).
- **TapeExecutor Terminal interface**: Decoupled from `ManagedTerminal` via interface with `Lock/Unlock/WritePty/GetTerminal/GetKeyEncoder/GetMouseEncoder/Resize/AddMarker`.
- **Golden file update mode**: Separate `update_golden` tool (not a flag on `verify_output`) to make accidental updates harder.
- **Scroll mouse actions**: `scroll_up`/`scroll_down` translated to button 4/5 press+release sequences.
- **All go-libghostty getters return `(value, error)`**: e.g., `CursorX() (uint16, error)`, `Cols() (uint16, error)`.
- **mcp-go CallToolRequest**: Use `req.GetString()`, `req.GetInt()`, `req.GetBool()`, `req.RequireString()`, `req.RequireInt()` — NOT `req.Params.Arguments["key"]`.
- **Output subscriber pattern**: `ManagedTerminal` broadcasts PTY output to subscribers via channels. `Subscribe()` returns an `outputSubscriber` with a `Channel()` method. Used by `attach` command to stream output without competing with `ptyReadLoop`.
- **Daemon protocol**: JSON-RPC over Unix socket. Each request is one JSON line, each response is one JSON line. `attach` is special: after the initial request, the daemon streams raw PTY output bytes to the client connection.
- **CLI output format**: Human-readable by default, `--json` flag for machine-parseable output.
- **Init-style subcommand registration**: Each subcommand file declares a package-level `*cobra.Command` var and registers it via `rootCmd.AddCommand()` in an `init()` function. No central list of subcommands in root.go.
- **Carapace completion actions**: Defined in `pkg/actions/` (`ActionTerminalIDs`, `ActionKeyNames`). Command files import and use them directly.
- **`run` command is daemon-free**: Creates terminal in-process, plays tape, verifies, exits. No socket needed. Always starts `sh` as the child process for tape input to work.

## Gotchas

- **libghostty-vt is built from git submodule**: Ghostty source is at `deps/ghostty` (pinned to `cb36966a752982014827a9cabcf630ec3788b3d9`). Run `make build-lib` to compile, or any `make build-*` target which depends on it.
- **Zig version**: System Zig 0.16.0 is incompatible. Use Zig 0.15.2 (set via `ZIG` make variable, defaults to `/tmp/zig-x86_64-linux-0.15.2/zig`).
- **Build command**: `make build-static` (builds libghostty-vt from submodule + compiles Go). Don't use bare `go build`.
- **`KeyUnknown` doesn't exist**: Use `KeyUnidentified` (value 0) as the zero/unknown sentinel for `libghostty.Key`.
- **`send_keys` accepts both PascalCase and snake_case key names**: `libghostty.ParseKey` only accepts snake_case (e.g. `tab`, `arrow_down`), but `SendKeys` normalizes W3C PascalCase names (e.g. `Tab`, `ArrowDown`) to snake_case automatically via `normalizeKey()`. Prefer PascalCase in MCP requests.
- **`MousePosition` uses `float32`**: Not `uint32` or `int`.
- **`Graphemes()` returns `[]uint32`**: Not a string — must convert codepoints to runes.
- **`CellWideWide`**: Not `CellWideYes`. Other constants: `CellWideNarrow`, `CellWideSpacerHead`, `CellWideSpacerTail`.
- **`StyleColor` is a struct**: With `Tag` (None/Palette/RGB), `Palette uint8`, `RGB ColorRGB`.
- **VHS parser API**: `parser.New(l)` takes lexer as argument; `p.Parse()` returns `[]Command` with no error return — check `p.Errors()` separately.
- **`go mod tidy` warnings**: The `go.mod` has indirect deps that gopls complains about — these come from mcp-go's transitive dependencies and are normal.
- **PTY bridge goroutines**: Exit when `done` channel is closed (in `Close()`). The `WritePty` callback runs synchronously during `VTWrite`.
- **No `MouseActionScrollUp/ScrollDown`**: go-libghostty only has Press/Release/Motion. Scroll is implemented via button 4/5 press+release.
- **`CursorStyle()` returns SGR text attributes, NOT cursor shape**: `Terminal.CursorStyle()` returns `*Style` (bold/italic/fg/bg), not the cursor shape. Use `RenderState.CursorVisualStyle()` instead: `NewRenderState()` → `rs.Update(term)` → `rs.CursorVisualStyle()` → compare against `CursorVisualStyleBar/Block/Underline/BlockHollow`. Always `defer rs.Close()`.
- **`ManagedTerminal.Close()` is idempotent**: Wrapped in `sync.Once` — safe to call multiple times.
- **`WaitFor` takes `sync.Locker`**: The first arg is the mutex, not the terminal. This allows per-iteration lock/unlock instead of holding the lock for the entire wait duration.
- **`List()` acquires per-terminal mutex**: Reads `Cols`/`Rows` under `mt.mu` to avoid data race with concurrent resize.
- **Don't read Ptmx from two goroutines**: `ptyReadLoop` already reads from `Ptmx`. To get PTY output, use `Subscribe()` instead of reading `Ptmx` directly — otherwise data is randomly split between readers.
- **`mcp` subcommand uses `server.New()`**: Not `daemon.New()`. The MCP server is a stdio JSON-RPC server; the daemon is a Unix socket server.
- **Daemon socket path**: `~/.local/share/tabdance-test/daemon.sock` (override with `TABDANCE_TEST_SOCK`). PID file at `~/.local/share/tabdance-test/daemon.pid`.
- **Daemon idle timeout**: 30s after last terminal closed and no client connected (override with `TABDANCE_TEST_IDLE_TIMEOUT`).

## Dependencies

| Package | Purpose |
|---------|---------|
| `go.mitchellh.com/libghostty` | Headless VT terminal, KeyEncoder, MouseEncoder, Formatter, GridRef |
| `github.com/mark3labs/mcp-go` | MCP server framework (tool registration, stdio transport) |
| `github.com/creack/pty` | PTY creation for child process spawning |
| `github.com/charmbracelet/vhs` | VHS tape parser (lexer, parser, token sub-packages) |
| `github.com/pmezard/go-difflib` | Unified diff generation for golden file comparison |
| `github.com/spf13/cobra` | CLI framework |
| `github.com/carapace-sh/carapace` | Shell completion integration |

## Testing

```sh
make test  # Builds libghostty-vt from submodule, then runs go test ./...
```

Integration tests in `internal/server/server_test.go` cover: create/close, list, get_view, get_cursor, get_size, resize, invalid terminal ID, and golden file verification for nonexistent files.

## Cross-Project Skills

- **ghostty** skill — go-libghostty API details (Terminal, KeyEncoder, MouseEncoder, Formatter, GridRef, etc.)
- **asciinema** skill — asciicast v2 format specification
- **vhs** skill — VHS tape DSL and parsing
- **bubbletea** skill — Bubble Tea framework internals
- **tabdance** skill — tabdance application internals
- **carapace** skill — carapace shell completion integration
