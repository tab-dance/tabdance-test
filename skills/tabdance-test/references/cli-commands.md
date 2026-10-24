# CLI Commands Reference

tabdance-test exposes a Cobra CLI with carapace shell completions. All commands (except `mcp` and `run`) communicate with a background daemon over a Unix socket.

## Daemon Architecture

- **Auto-start**: Daemon forks automatically on first CLI command if not running
- **Socket**: `~/.local/share/tabdance-test/daemon.sock` (override: `TABDANCE_TEST_SOCK`)
- **PID file**: `~/.local/share/tabdance-test/daemon.pid`
- **Idle timeout**: 30s after last terminal closed and no client connected (override: `TABDANCE_TEST_IDLE_TIMEOUT`)
- **Protocol**: JSON-RPC over Unix socket (one JSON line per request/response)
- **`run` command is daemon-free**: Creates terminal in-process, no socket needed

## Command Reference

### Terminal Lifecycle

| Command | Usage | Description |
|---------|-------|-------------|
| `exec` | `tabdance-test exec <command> [args...]` | Create terminal running command, print ID (use `attach <id>` to stream output) |
| `ls` | `tabdance-test ls` | List active terminals |
| `inspect` | `tabdance-test inspect <id>` | Terminal details (size, command, recording status) |
| `attach` | `tabdance-test attach <id>` | Stream PTY output (Ctrl+C to detach) |
| `kill` | `tabdance-test kill <id>` | Hidden alias for `rm` (deprecated) |
| `rm` | `tabdance-test rm <id>` | Remove/close terminal |

### `exec` flags

| Flag | Default | Description |
|------|---------|-------------|
| `--cols` | 108 | Terminal width |
| `--rows` | 24 | Terminal height |
| `--scrollback` | 10000 | Max scrollback lines |
| `--env` | (none) | Environment variables (key=value, repeatable) |
| `--working-directory` | (none) | Working directory for child process |
| `--record` | (none) | Start asciinema recording to path |

### Inspection

| Command | Usage | Description |
|---------|-------|-------------|
| `view` | `tabdance-test view <id>` | Print screen content |
| `cursor` | `tabdance-test cursor <id>` | Print cursor position and state |
| `size` | `tabdance-test size <id>` | Print terminal dimensions |
| `cell` | `tabdance-test cell <id> <x> <y>` | Inspect specific cell |

All inspection commands support `--json` for machine-parseable output.

### `view` flags

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | plain | Output format: plain, vt, html |
| `--trim` | true | Trim trailing whitespace |
| `--unwrap` | true | Unwrap soft-wrapped lines |
| `--json` | false | Output as JSON |

### Input

| Command | Usage | Description |
|---------|-------|-------------|
| `send` | `tabdance-test send <id> <text>` | Send raw text to PTY stdin |
| `keys` | `tabdance-test keys <id> <key...>` | Send key events (e.g. Enter, Tab, Ctrl+C) |
| `mouse` | `tabdance-test mouse <id> <action> <button> <x> <y>` | Send mouse event |
| `paste` | `tabdance-test paste <id> <text>` | Send paste event (bracketed-paste-aware) |
| `focus` | `tabdance-test focus <id> <gained\|lost>` | Send focus event |
| `resize` | `tabdance-test resize <id> <cols> <rows>` | Resize terminal |

### `mouse` actions

| Action | Description |
|--------|-------------|
| `press` | Mouse button press |
| `release` | Mouse button release |
| `move` | Mouse movement |
| `scroll_up` | Scroll up (button 4) |
| `scroll_down` | Scroll down (button 5) |

### `mouse` buttons

`left`, `middle`, `right`

### `keys` key names (PascalCase)

`Tab`, `Enter`, `Escape`, `Backspace`, `Delete`, `Insert`, `Home`, `End`, `PageUp`, `PageDown`, `Space`, `ArrowUp`, `ArrowDown`, `ArrowLeft`, `ArrowRight`, `F1`–`F12`, `Ctrl+A`–`Ctrl+Z`

### Tape & Recording

| Command | Usage | Description |
|---------|-------|-------------|
| `tape` | `tabdance-test tape <id> <path>` | Play .tape file in terminal |
| `record` | `tabdance-test record <id> <path>` | Start asciinema recording (`record <id> --stop` stops it) |
| `marker` | `tabdance-test marker <id> <label>` | Add recording marker |

### `tape` flags

| Flag | Default | Description |
|------|---------|-------------|
| `--timeout` | 30 | Timeout in seconds for Wait commands |

### `record` flags

| Flag | Default | Description |
|------|---------|-------------|
| `--idle-time-limit` | 2.0 | Idle time limit in seconds |

### Golden Files

| Command | Usage | Description |
|---------|-------|-------------|
| `verify` | `tabdance-test verify <id> <golden>` | Verify output against golden file |
| `snapshot` | `tabdance-test snapshot <id> <golden>` | Update/create golden file |

Both support `--format`, `--trim`, `--unwrap` flags (same as `view`).

### Wait

| Command | Usage | Description |
|---------|-------|-------------|
| `wait` | `tabdance-test wait <id> <pattern>` | Wait for regex pattern on screen |

### `wait` flags

| Flag | Default | Description |
|------|---------|-------------|
| `--scope` | screen | Match scope: screen or line |
| `--timeout` | 10 | Timeout in seconds |
| `--format` | plain | Output format |
| `--trim` | true | Trim trailing whitespace |
| `--unwrap` | true | Unwrap soft-wrapped lines |

### Run (Daemon-Free)

| Command | Usage | Description |
|---------|-------|-------------|
| `run` | `tabdance-test run <tape>` | Play tape, verify golden, exit 0/1 |

### `run` flags

| Flag | Default | Description |
|------|---------|-------------|
| `--golden` | (none) | Golden file to verify against |
| `--update` | false | Update golden file instead of verifying |
| `--cols` | 108 | Terminal width |
| `--rows` | 24 | Terminal height |
| `--timeout` | 30 | Timeout in seconds for Wait commands |
| `--record` | (none) | Record session to .cast file |
| `--env` | (none) | Environment variables (key=value, repeatable) |
| `--working-directory` | (none) | Working directory |

### MCP

| Command | Usage | Description |
|---------|-------|-------------|
| `mcp` | `tabdance-test mcp` | Start MCP server (stdio) |

## Output Formats

- **Human-readable** (default): tables, plain text, key=value pairs
- **`--json`**: JSON objects for scripting (pipe into `jq`)

### `ls` human-readable output

```
ID       COLS  ROWS  COMMAND
term_1   108   24    bash
term_2   80    24    vim test.go
```

### `view` human-readable output

Plain text screen content (no metadata).

### `cursor` human-readable output

```
x:5 y:2 visible:true style:bar pending_wrap:false
```

### `size` human-readable output

```
108x24 (864x384px)
```

### `cell` human-readable output

```
codepoint:65 text:"A" wide:false bold:false italic:false fg:#ffffff bg:#000000
```

## Carapace Completions

Dynamic completions are registered for:

- **Terminal IDs**: All commands taking `<id>` query the daemon for active terminals
- **Key names**: `keys` command completes PascalCase key names
- **Mouse actions/buttons**: `mouse` command
- **Format values**: `view`, `verify`, `snapshot`, `wait` complete `plain`/`vt`/`html`
- **Scope values**: `wait` completes `screen`/`line`
- **Focus events**: `focus` completes `gained`/`lost`
- **File paths**: `.tape` files for `tape`/`run`, `.cast` for `record`, any file for `verify`/`snapshot`
- **Executables**: `exec` first positional argument
