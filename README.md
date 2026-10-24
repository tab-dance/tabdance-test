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
tabdance-test exec bash
# e.g. eager_borg

# Inspect it
tabdance-test ls --json
tabdance-test view eager_borg
tabdance-test cursor eager_borg --json

# Drive it
tabdance-test send eager_borg "echo hello world"
tabdance-test keys eager_borg Enter
tabdance-test wait eager_borg "hello world"

# Attach to the live PTY stream (Ctrl+C to detach)
tabdance-test attach eager_borg

# Clean up
tabdance-test kill eager_borg
```

Daemon-backed commands (`exec`, `view`, `keys`, ...) auto-start a background
daemon on first use. It idles out 30s after the last terminal is closed.
Override with `TABDANCE_TEST_SOCK` and `TABDANCE_TEST_IDLE_TIMEOUT`.

### Golden file testing (CI-friendly)

```sh
# Play a VHS tape and verify against a .golden file; exits 0/1
tabdance-test run tapes/echo.tape --golden tapes/echo.golden

# Run every tape in a directory, each against its matching .golden
tabdance-test run tapes/

# Create/update a golden file from the current terminal state
tabdance-test snapshot eager_borg golden.txt
tabdance-test view eager_borg --format json > golden.json
```

Tapes support `Set Shell`, `Set Width`, `Set Height`, `Set WaitTimeout`,
`Set TypingSpeed`, `Wait /regex/` and full VHS typing syntax.

### MCP mode

```sh
tabdance-test mcp
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
tabdance-test exec bash --record session.cast        # start recording at creation
tabdance-test record eager_borg session.cast          # or start later
tabdance-test marker eager_borg "after login"
tabdance-test stop eager_borg
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
| Record | `record`, `stop`, `marker` | `start_recording`, `stop_recording`, `add_marker` |

## Commands

Run `tabdance-test --help` for the full list.

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
