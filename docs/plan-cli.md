# tabdance-test CLI — Implementation Plan

## Overview

Refactor tabdance-test from an MCP-only binary to a Cobra CLI with carapace completions. The MCP server becomes a subcommand. All MCP features are exposed as CLI commands for local development and CI testing.

## Command Tree

```
tabdance-test
├── mcp                          # MCP server (stdio) — existing behavior
├── exec <command>               # Create terminal running command, print ID
│   --cols 108                   # Terminal width
│   --rows 24                    # Terminal height
│   --scrollback 10000           # Max scrollback
│   --env <key=val>              # Environment variables (repeatable)
│   --working-directory <dir>    # Working directory
│   --record <path>              # Start recording immediately
│   --detach / -d                # Daemonize (default: attach to PTY output)
├── ls                           # List active terminals
├── inspect <id>                 # Inspect terminal details (size, command, etc.)
├── attach <id>                  # Attach to terminal's PTY output (streaming)
├── kill <id>                    # Kill terminal
├── rm <id>                      # Remove/close terminal
├── view <id>                    # Print screen content
│   --format plain|vt|html
│   --trim / --no-trim
│   --unwrap / --no-unwrap
│   --json                       # JSON output
├── cursor <id>                  # Print cursor state
│   --json
├── size <id>                    # Print dimensions
│   --json
├── cell <id> <x> <y>            # Inspect specific cell
│   --json
├── send <id> <text>             # Send raw text
├── keys <id> <key...>           # Send key events
├── mouse <id> <action> <button> <x> <y>
├── paste <id> <text>            # Send paste event
├── focus <id> <gained|lost>     # Send focus event
├── resize <id> <cols> <rows>    # Resize terminal
├── tape <id> <path>             # Play .tape file into terminal
│   --timeout 30
├── wait <id> <pattern>          # Wait for regex on screen
│   --scope screen|line
│   --timeout 10
├── verify <id> <golden>         # Verify output against golden file
│   --format / --trim / --unwrap
├── snapshot <id> <golden>       # Update/create golden file
│   --format / --trim / --unwrap
├── record <id> <path>           # Start asciinema recording
│   --idle-time-limit 2.0
├── stop <id>                    # Stop recording
├── marker <id> <label>          # Add recording marker
└── run <tape>                   # One-shot: create terminal, play tape, verify, exit 0/1
    --golden <path>
    --update                     # Update golden instead of verifying
    --cols 108 / --rows 24
    --timeout 30
    --record <path>
    --env <key=val>
    --working-directory <dir>
```

## Daemon Architecture

### Overview

The CLI commands (except `mcp` and `run`) communicate with a background daemon process over a Unix socket. The daemon holds the `TerminalManager` and all active terminals.

```
CLI client                    Daemon
  ┌──────────┐               ┌──────────────────────┐
  │ tabdance- │  Unix socket  │  TerminalManager      │
  │ test view │ ────────────► │  ├─ term_1 (bash)    │
  │           │  JSON-RPC     │  ├─ term_2 (vim)     │
  │           │ ◄──────────── │  └─ term_3 (htop)    │
  └──────────┘               └──────────────────────┘
```

### Daemon Lifecycle

| Scenario | Behavior |
|----------|----------|
| No daemon running, any command | Auto-start daemon, then connect |
| `exec -d bash` | Create terminal, print ID, return to shell |
| `exec bash` (no -d) | Create terminal, attach to PTY output (streaming) |
| `view term_1` | Connect to daemon, get screen, print, disconnect |
| Last terminal closed | Daemon idles 30s, then exits |
| `run tape.tape` | No daemon — in-process, standalone |
| `SIGTERM` to daemon | Close all terminals, clean up socket, exit |

### Socket Path

- Default: `~/.local/share/tabdance-test/daemon.sock`
- Override: `TABDANCE_TEST_SOCK` env var
- PID file: `~/.local/share/tabdance-test/daemon.pid`
- Idle timeout: 30s after last terminal closed and no client connected
- Override: `TABDANCE_TEST_IDLE_TIMEOUT` env var

### Auto-Start Behavior

When any CLI command is invoked and no daemon is running:
1. Fork daemon process (`exec` into background)
2. Wait for socket to appear (poll with timeout)
3. Connect and proceed

### Protocol

JSON-RPC over Unix socket (simple request/response). Each request is a single JSON line, each response is a single JSON line.

**Request:**
```json
{"method": "view", "params": {"id": "term_1", "format": "plain", "trim": true, "unwrap": true}}
```

**Response (success):**
```json
{"ok": true, "result": {"content": "...", "cursor_x": 5, "cursor_y": 2}}
```

**Response (error):**
```json
{"ok": false, "error": "terminal \"term_1\" not found"}
```

### Streaming (attach)

`attach` and `exec` (without `-d`) need streaming PTY output. This uses a separate connection mode:
- Client connects to socket
- Sends `{"method": "attach", "params": {"id": "term_1"}}`
- Daemon streams PTY output frames as they arrive
- Client proxies raw bytes to its own stdout
- Ctrl+C sends detach request, daemon stops streaming

## `run` Command (Daemon-Free)

For CI and scripting. Creates a terminal in-process, plays the tape, optionally verifies against a golden file, and exits.

**Flow:**
1. Create terminal with given cols/rows/env/working-directory
2. Play tape file
3. If `--golden` provided:
   - Verify output against golden file
   - On match: print `PASS` to stderr, exit 0
   - On mismatch: print `FAIL` + diff to stderr, exit 1
4. If `--update` provided:
   - Write current output to golden file
   - Print `UPDATED` to stderr, exit 0
5. If no `--golden`: exit 0

No daemon involved. No socket. Pure in-process.

## Output Format

All inspect commands default to human-readable output. `--json` flag switches to JSON.

**`ls` (human-readable):**
```
ID       COLS  ROWS  COMMAND
term_1   108   24    bash
term_2   80    24    vim test.go
```

**`ls` (--json):**
```json
[{"id": "term_1", "cols": 108, "rows": 24, "command": "bash"}]
```

**`view` (human-readable):** plain text screen content
**`view` (--json):** `{"content": "...", "cursor_x": 5, "cursor_y": 2, "cursor_visible": true, "cursor_style": "bar", "active_screen": "main"}`

**`cursor` (human-readable):** `x:5 y:2 visible:true style:bar pending_wrap:false`
**`cursor` (--json):** `{"x": 5, "y": 2, "visible": true, "style": "bar", "pending_wrap": false}`

**`size` (human-readable):** `108x24 (864x384px)`
**`size` (--json):** `{"cols": 108, "rows": 24, "width_px": 864, "height_px": 384}`

**`cell` (human-readable):** `codepoint:A wide:false bold:false italic:false fg:#ffffff bg:#000000`
**`cell` (--json):** `{"codepoint": "A", "text": "A", "wide": false, "bold": false, "italic": false, "fg": "#ffffff", "bg": "#000000", "hyperlink": null}`

**`inspect` (human-readable):**
```
ID:       term_1
Command:  bash
Size:     108x24
Scrollback: 10000
Recording: /tmp/session.cast
```

## Carapace Completions

### Flag Completions

| Command | Flag | Completion |
|---------|------|------------|
| `exec` | `--format` | `plain`, `vt`, `html` |
| `view` | `--format` | `plain`, `vt`, `html` |
| `view` | `--trim` / `--no-trim` | (boolean flag, no completion needed) |
| `wait` | `--scope` | `screen`, `line` |
| `focus` | positional | `gained`, `lost` |
| `mouse` | action | `press`, `release`, `move` |
| `mouse` | button | `left`, `middle`, `right` |
| `record` | `--idle-time-limit` | (numeric, no completion) |
| `tape` | `--timeout` | (numeric, no completion) |
| `wait` | `--timeout` | (numeric, no completion) |

### Positional Completions

| Command | Position | Completion |
|---------|----------|------------|
| `view`, `cursor`, `size`, `cell`, `send`, `keys`, `mouse`, `paste`, `focus`, `resize`, `tape`, `wait`, `verify`, `snapshot`, `record`, `stop`, `marker`, `attach`, `kill`, `rm`, `inspect` | 1 (terminal ID) | Dynamic: query daemon for active terminal IDs |
| `cell` | 2, 3 (x, y) | (numeric, no completion) |
| `mouse` | 3, 4 (x, y) | (numeric, no completion) |
| `resize` | 2, 3 (cols, rows) | (numeric, no completion) |
| `keys` | 2+ (key names) | `Tab`, `Enter`, `Escape`, `ArrowDown`, `ArrowUp`, `ArrowLeft`, `ArrowRight`, `Backspace`, `Delete`, `Insert`, `Home`, `End`, `PageUp`, `PageDown`, `Space`, `Ctrl+A`..`Ctrl+Z`, `F1`..`F12` |
| `tape`, `run` | tape path | `ActionFiles(".tape")` |
| `verify`, `snapshot` | golden path | `ActionFiles()` |
| `record` | output path | `ActionFiles(".cast")` |
| `exec` | command | `ActionExecutables()` |

### Terminal ID Completion

Dynamic completion that queries the daemon for active terminals. Falls back to empty list if daemon is not running (graceful degradation, no error).

```go
carapace.ActionCallback(func(c carapace.Context) carapace.Action {
    // Query daemon for terminal list
    // Return ActionValues(term_1, term_2, ...)
})
```

## Project Structure

```
cmd/
  tabdance-test/
    main.go                    # Entrypoint: root command, execute

internal/
  cli/
    root.go                    # Root command + cobra setup + carapace integration
    mcp.go                     # `mcp` subcommand
    exec.go                    # `exec` subcommand
    ls.go                      # `ls` subcommand
    inspect.go                 # `inspect` subcommand
    attach.go                  # `attach` subcommand
    kill.go                    # `kill` subcommand
    rm.go                      # `rm` subcommand
    view.go                    # `view` subcommand
    cursor.go                  # `cursor` subcommand
    size.go                    # `size` subcommand
    cell.go                    # `cell` subcommand
    send.go                    # `send` subcommand
    keys.go                    # `keys` subcommand
    mouse.go                   # `mouse` subcommand
    paste.go                   # `paste` subcommand
    focus.go                   # `focus` subcommand
    resize.go                  # `resize` subcommand
    tape.go                    # `tape` subcommand
    wait.go                    # `wait` subcommand
    verify.go                  # `verify` subcommand
    snapshot.go                # `snapshot` subcommand
    record.go                  # `record` subcommand
    stop.go                    # `stop` subcommand
    marker.go                  # `marker` subcommand
    run.go                     # `run` subcommand (daemon-free)
    completion.go              # Carapace completion helpers (terminal IDs, key names, etc.)
  daemon/
    daemon.go                  # Daemon process: socket listener, request router
    protocol.go                # JSON-RPC protocol types (Request, Response)
    handler.go                 # Request handler: dispatches to TerminalManager ops
  client/
    client.go                  # Daemon client: connect, send request, receive response
    attach.go                  # Streaming attach client
  server/                      # (existing) MCP server — unchanged
    server.go
    server_test.go
  terminal/                    # (existing) Terminal management — unchanged
    manager.go
    managed.go
  input/                       # (existing) Input injection — unchanged
    keys.go
    mouse.go
    focus.go
    paste.go
  inspect/                     # (existing) Terminal inspection — unchanged
    view.go
    cell.go
    cursor.go
    size.go
  recording/                   # (existing) Asciicast recording — unchanged
    asciicast.go
  tape/                        # (existing) VHS tape playback — unchanged
    executor.go
  golden/                      # (existing) Golden file verification — unchanged
    verify.go
  wait/                        # (existing) Wait/poll — unchanged
    wait.go
```

## Dependencies (New)

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/carapace-sh/carapace` | Shell completion integration |
| `github.com/carapace-sh/carapace/pkg/style` | Completion styling |

## Implementation Steps

### Step 1: Cobra + Carapace Scaffolding

- Add cobra and carapace dependencies
- Create `internal/cli/root.go` with root command, carapace integration
- Update `cmd/tabdance-test/main.go` to use root command
- Create stub files for all subcommands
- Verify: `make build-static && ./tabdance-test --help`

### Step 2: MCP Subcommand

- Move current `server.New().Run()` into `internal/cli/mcp.go`
- Verify: `./tabdance-test mcp` works as before

### Step 3: Daemon + Client

- Create `internal/daemon/protocol.go` — JSON-RPC request/response types
- Create `internal/daemon/daemon.go` — socket listener, accept loop, request routing
- Create `internal/daemon/handler.go` — dispatch requests to TerminalManager
- Create `internal/client/client.go` — connect to socket, send/receive
- Daemon auto-start logic in client
- Idle timeout for daemon
- Verify: start daemon, connect, create terminal, list, close

### Step 4: Terminal Lifecycle Commands

- `exec` — create terminal (with/without detach)
- `ls` — list terminals
- `inspect` — terminal details
- `kill` — kill terminal
- `rm` — close and remove terminal
- Verify: `./tabdance-test exec -d bash && ./tabdance-test ls && ./tabdance-test rm term_1`

### Step 5: Attach Command

- `attach` — streaming PTY output
- `exec` without `-d` — create + attach
- Ctrl+C detach behavior
- Verify: `./tabdance-test exec bash` shows live terminal output

### Step 6: Inspection Commands

- `view`, `cursor`, `size`, `cell`
- Human-readable + `--json` output
- Verify: create terminal, write content, inspect

### Step 7: Input Commands

- `send`, `keys`, `mouse`, `paste`, `focus`
- Verify: create terminal with bash, send keys, view result

### Step 8: Tape + Recording Commands

- `tape` — play .tape file
- `record`, `stop`, `marker`
- Verify: record a session, play a tape

### Step 9: Golden File Commands

- `verify`, `snapshot`
- Verify: create terminal, snapshot, modify, verify

### Step 10: Wait Command

- `wait` — regex polling
- Verify: create terminal, start slow command, wait for pattern

### Step 11: Run Command (Daemon-Free)

- `run` — in-process tape execution + golden verification
- Exit codes: 0 (pass), 1 (fail)
- Stderr output: PASS / FAIL + diff
- Verify: `./tabdance-test run test.tape --golden test.golden`

### Step 12: Carapace Completions

- Terminal ID dynamic completion (query daemon)
- Key name completion
- Mouse action/button completion
- Format/scope value completion
- File path completions (.tape, .cast, golden files)
- Verify: completion works in bash/zsh/fish

### Step 13: Resize Command

- `resize` — resize terminal + encoder re-sync
- Verify: resize and check new dimensions

## Open Questions

- **Attach streaming format**: Raw VT output (see real TUI) vs formatted plain text? Current plan: raw by default, `--format plain` for piping.
- **Daemon auto-start reliability**: Edge cases around race conditions (two clients starting daemon simultaneously). Need file locking on PID file.
- **`exec` without `-d` and Ctrl+C**: Should Ctrl+C detach (keep terminal alive) or kill the terminal? Current plan: detach only.
