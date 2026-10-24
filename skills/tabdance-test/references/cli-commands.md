# CLI Commands Reference

tabdance-test (tt) exposes a Cobra CLI with carapace shell completions. All commands (except `mcp` and `run`) communicate with a background daemon over a Unix socket.

## Daemon Architecture

- **Auto-start**: Daemon forks automatically on first CLI command if not running
- **Socket**: `~/.local/share/tabdance-test/<project-key>/daemon.sock`, where `<project-key>` hashes the git repo root (or cwd); terminals never bleed between projects (override: `TABDANCE_TEST_SOCK`); dir 0700, socket 0600
- **PID file**: `~/.local/share/tabdance-test/daemon.pid`, plus `daemon.pid.lock` (flock, prevents start races)
- **Idle timeout**: 30s after last terminal closed and no client connected (override: `TABDANCE_TEST_IDLE_TIMEOUT`)
- **Shutdown**: SIGINT/SIGTERM remove socket + pid files and close all terminals
- **`.term`**: `exec --term`/`run --term` override the child's TERM
- **`cast2tape`**: converts recordings into deterministic tapes (input events -> Type/Key/Ctrl/Raw, WaitStable between input clusters)
- **Protocol**: JSON-RPC over Unix socket (one JSON line per request/response)
- **`run` command is daemon-free**: Creates terminal in-process, no socket needed

## Command Reference

### Terminal Lifecycle

| Command | Usage | Description |
|---------|-------|-------------|
| `exec` | `tt exec <command> [args...]` | Create terminal running command, print ID (use `attach <id>` to stream output) |
| `ls` | `tt ls` | List active terminals |
| `inspect` | `tt inspect <id>` | Terminal details (size, command, recording, child exit status) |
| `output` | `tt output <id> [bytes]` | Print recent raw PTY output (diagnostics) |
| `doctor` | `tt doctor` | Environment preflight checks (spawns nothing) |
| `attach` | `tt attach <id>` | Stream PTY output (Ctrl+C to detach) |
| `kill` | `tt kill <id>` | Hidden alias for `rm` (deprecated) |
| `rm` | `tt rm <id>` | Remove/close terminal |

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
| `view` | `tt view <id>` | Print screen content (`--format` explicit wins; piped default plain, interactive default vt) |
| `cursor` | `tt cursor <id>` | Print cursor position and state |
| `size` | `tt size <id>` | Print terminal dimensions |
| `cell` | `tt cell <id> <x> <y>` | Inspect specific cell |

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
| `send` | `tt send <id> <text>` | Send raw text to PTY stdin |
| `keys` | `tt keys <id> <key...>` | Send key events (e.g. Enter, Tab, Ctrl+C) |
| `mouse` | `tt mouse <id> <action> <button> <x> <y>` | Send mouse event |
| `paste` | `tt paste <id> <text>` | Send paste event (bracketed-paste-aware) |
| `focus` | `tt focus <id> <gained\|lost>` | Send focus event |
| `resize` | `tt resize <id> <cols> <rows>` | Resize terminal |

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
| `tape` | `tt tape <id> <path>` | Play .tape file in terminal |
| `record` | `tt record <id> <path>` | Start asciinema recording (`record <id> --stop` stops it) |
| `marker` | `tt marker <id> <label>` | Add recording marker |

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
| `verify` | `tt verify <id> <golden>` | Verify output against golden file |
| `snapshot` | `tt snapshot <id> <golden>` | Update/create golden file |

Both support `--format`, `--trim`, `--unwrap` flags (same as `view`).

### Wait

| Command | Usage | Description |
|---------|-------|-------------|
| `wait` | `tt wait <id> <pattern>` | Wait for regex pattern on screen |

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
| `run` | `tt run <tape>` | Play tape, verify golden, exit 0/1 |

### `run` flags

| Flag | Default | Description |
|------|---------|-------------|
| `--golden` | (none) | Golden file to verify against |
| `--update` | false | Update golden file instead of verifying |
| `--cols` | 108 | Terminal width |
| `--rows` | 24 | Terminal height |
| `--scrollback` | 10000 | Max scrollback lines |
| `--timeout` | 30 | Timeout in seconds for Wait commands |
| `--record` | (none) | Record session to .cast file |
| `--env` | (none) | Environment variables (key=value, repeatable) |
| `--working-directory` | (none) | Working directory |
| `--json` | false | Directory runs: print {passed,failed,skipped} JSON summary |
| `--strict` | false | Reject tapes containing Sleep verbs |

Directory receipts: running a directory verifies each `*.tape` against its
matching `.golden`; without `--update`, tapes that have no golden yet are
listed under SKIPPED. Exit code is 1 if any tape failed.

### MCP

| Command | Usage | Description |
|---------|-------|-------------|
| `mcp` | `tt mcp` | Start MCP server (stdio) |

## Output Formats

- **Human-readable** (default): tables, plain text, key=value pairs
- **`--json`**: JSON objects for scripting (pipe into `jq`); available on `ls`, `inspect`, `view`, `cursor`, `size`, `cell`, `wait`, `verify`, and directory `run`

### `ls` human-readable output

```
ID                  COLS  ROWS  RECORDING  COMMAND
bold_sutherland     108   24    yes        bash
reverent_hertz      80    24    -          vim test.go
```

Terminal IDs are collision-checked random adjective+noun names; you can
always inspect `ls --json` for the authoritative list.

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
- **Format values**: `view`/`verify`/`snapshot` complete `plain`/`vt`/`html`/`svg`/`json`; `wait` completes `plain`/`vt`/`html` (its poller only reads rendered text)
- **Scope values**: `wait` completes `screen`/`line`
- **Focus events**: `focus` completes `gained`/`lost`
- **File paths**: `.tape` files for `tape`/`run`, `.cast` for `record`, any file for `verify`/`snapshot`
- **Executables**: `exec` first positional argument
