# tabdance-test

Headless terminal testing with [go-libghostty](https://go.mitchellh.com/libghostty).

tabdance-test spins up real terminal emulators (the full Ghostty VT engine,
statically linked, no display needed), runs command output and TUI apps inside
them, and gives you precise, scriptable control: inject keys and mouse events,
read back the rendered screen (plain text, VT, HTML, SVG, JSON grid), wait for
patterns, play [VHS](https://github.com/charmbracelet/vhs) tapes, and verify
output against golden files. Perfect for testing shells, completers, and
Bubble Tea/charm applications end-to-end.

Two interfaces, same capabilities:

- **CLI** (Cobra + carapace completions) — for local development and CI
  (completions install via [carapace](https://github.com/carapace-sh/carapace-bin); the cobra-built-in `completion`/`help` commands are disabled to keep the surface honest)
- **MCP server** (stdio) — for AI agent-driven testing

## Install

```sh
# from source (builds libghostty-vt from the deps/ghostty submodule)
make build-static

# release archives are published for GitHub releases (linux amd64/arm64)
```

Requirements: Go 1.26+, Zig, and the `deps/ghostty` git submodule
(`git submodule update --init --recursive`).

## Quickstart

```sh
# Create a terminal running a command and print its ID
tt exec bash
# e.g. eager_borg

# Inspect it (add --format to force; piped output is plain,
# a terminal run replays the screen with colors as vt)
tt ls --json
tt view eager_borg
tt cursor eager_borg --json

# Drive it
tt send eager_borg "echo hello world"
tt keys eager_borg Enter
tt wait eager_borg "hello world"

# Attach to the live PTY stream (Ctrl+C to detach)
tt attach eager_borg

# Clean up
tt kill eager_borg
```

Daemon-backed commands (`exec`, `view`, `keys`, ...) auto-start a background
daemon on first use — one shared daemon for all projects. Terminals carry
the tag of their creating project, so `tt ls` lists everything (PROJECT
column) and terminal-ID completion stays scoped to your project; agent
flows that must only see their own project use `tt ls --mine`. Direct-ID
commands work across projects. The daemon idles out 30s after the last
terminal is closed (disable with `TABDANCE_TEST_IDLE_TIMEOUT=0`); the
socket directory can be moved with `TABDANCE_TEST_SOCK`.

### Golden file testing (CI-friendly)

```sh
# Play a VHS tape and verify against a .golden file; exits 0/1
tt run tapes/echo.tape --golden tapes/echo.golden

# Run every tape in a directory, each against its matching .golden
tt run tapes/

# Create/update a golden file from the current terminal state
tt snapshot eager_borg golden.txt
tt view eager_borg --format json > golden.json
```

Tapes support `Set Shell`, `Set Width`, `Set Height`, `Set WaitTimeout`,
`Set TypingSpeed`, `Wait /regex/` and full VHS typing syntax.

### MCP mode

```sh
tt mcp
```

Connect your MCP client to this stdio server. Available tools: `create_terminal`,
`list_terminals`, `close_terminal`, `reset_terminal`, `resize`, `send_keys`,
`send_text`, `send_mouse`, `send_focus`, `send_paste`, `get_view`, `get_cell`,
`get_cursor`, `get_size`, `start_recording`, `stop_recording`, `add_marker`,
`play_tape`, `verify_output`, `update_golden`, `wait_for`.

The MCP server operates on the same headless terminal primitives as the CLI,
including input into and output from interactive TUI applications such as
those built with charmbracelet/bubbletea.

### Recording

```sh
tt exec bash --record session.cast        # start recording at creation
tt record eager_borg session.cast          # or start later
tt marker eager_borg "after login"
tt record eager_borg --stop
```

Recordings are [asciicast v2](https://docs.asciinema.org/manual/asciicast/v2/)
files, directly playable with `asciinema play`.

## Feature matrix

| Capability | CLI | MCP |
|-----------------|----------------|---------------------|
| Create/destroy terminals | `exec`, `kill`, `rm` | `create_terminal`, `close_terminal` |
| Send input | `send`, `keys`, `mouse`, `paste`, `focus`, `resize` | `send_text`, `send_keys`, `send_mouse`, `send_paste`, `send_focus`, `resize` |
| Inspect state | `view`, `cursor`, `size`, `cell`, `inspect`, `ls` | `get_view`, `get_cursor`, `get_size`, `get_cell`, `list_terminals` |
| Assert behavior | `wait`, `verify`, `run` | `wait_for`, `verify_output`, `play_tape` |
| Record | `record` (`--stop`), `marker` | `start_recording`, `stop_recording`, `add_marker` |

## Commands

Run `tt --help` for the full list.

## Development

```sh
make build-static   # build libghostty-vt + statically link (default)
make build-dynamic  # quick iteration without full static link
make test           # build + run all tests
make test-race      # tests with the race detector
make fmt            # gofmt all packages
make vet            # go vet
```

Architecture overview: see [AGENTS.md](AGENTS.md).
