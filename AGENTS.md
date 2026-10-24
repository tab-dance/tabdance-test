# AGENTS.md

## Project Overview

tabdance-test is a CLI tool and MCP server for headless terminal testing using [go-libghostty](https://go.mitchellh.com/libghostty). It exposes two interfaces:

1. **CLI** (Cobra + carapace completions) — for local development and CI testing
2. **MCP server** (stdio) — for AI agent–driven testing

Both interfaces provide the same capabilities: create terminals, send keys/mouse, inspect state, record sessions, play VHS tapes, and verify output against golden files.

## Essential Commands

```sh
make build-static           # Build libghostty-vt + static link (default)
make build-dynamic          # Build libghostty-vt + dynamic link
make build-lib              # Build libghostty-vt from submodule only
make test                   # Build + run tests
make test-race              # Build + run tests with race detector
make fmt                    # Run gofmt over all Go packages
make vet                    # Run go vet
make clean                  # Remove binary
make run                    # Build dynamic + run
```

Ghostty is included as a git submodule at `deps/ghostty` (tracking Ghostty `main`). The `build-lib` target compiles `libghostty-vt` from the submodule using Zig; it is skipped when `deps/ghostty/zig-out/lib/libghostty-vt.a` already exists, so incremental `make test` runs don't recompile the library. All other build targets depend on `build-lib` automatically. Requires Zig 0.17.0 or newer on the PATH (override with the `ZIG` make variable).

## Releasing

Pushing a `v*` tag triggers `.github/workflows/release.yml`, which runs goreleaser on ubuntu-24.04 to build static Linux binaries (amd64 + arm64) and publish a draft GitHub release with tar.gz archives and checksums. Key points:

- `.goreleaser.yaml` uses the Go builder with a `hooks.pre` script (`scripts/build-lib.sh <target>`) that cross-compiles `libghostty-vt` per target into `dist/lib/<target>` with `zig build -Dtarget=...`. Each target gets its own pkg-config tree so parallel builds don't clobber each other.
- `PKG_CONFIG_PATH` must be absolute (cgo runs pkg-config from the go.mitchellh.com/libghostty module directory), hence the `GITHUB_WORKSPACE` template.
- Cross-linking uses `zig cc -target <triple>` as `CC` per target (see `overrides`).
- Local dry run: `GITHUB_WORKSPACE=$PWD ZIG=$(command -v zig) goreleaser release --snapshot --clean`
- The build injects `-X internal/version.Version={{ .Version }}` so `tabdance-test --version` matches the release tag.
- `.github/workflows/build.yml` runs the same goreleaser snapshot build (plus `goreleaser check`) on pull requests and pushes to master, so release builds are verified before tagging.
- Only Linux targets are currently built; macOS/Windows need per-platform runners (and, for macOS, an SDK).

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

1. CLI command invoked (e.g. `tabdance-test view eager_borg`)
2. Client checks if daemon is running (PID file + socket connect)
3. If not: fork daemon process (`tabdance-test --daemon`), wait for socket
4. Connect, send JSON-RPC request, receive response, disconnect
5. Daemon idles out after 30s with no active terminals or clients
6. If two clients race to auto-start, `daemon.lock` (flock) makes the loser exit instead of stealing the winner's socket file

### Key flow: create terminal with command

1. `libghostty.NewTerminal()` with `WithSize`, `WithMaxScrollback`, `WithWritePty`
2. Create `KeyEncoder` + `MouseEncoder`, sync with `SetOptFromTerminal`
3. If command provided: spawn PTY child process, start bridge goroutines; a dedicated goroutine calls `cmd.Wait()` to reap the child so it cannot haunt the daemon as a zombie
4. Bridge: child stdout → `term.VTWrite()` + optional `recorder.WriteOutput()` + broadcast to subscribers
5. Bridge: `WritePty` callback → atomic PTY pointer load → `ptmx.Write()` → child stdin (responses to shell queries reach the child synchronously during VTWrite)

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
│   --cols / --rows / --scrollback / --env / --working-directory / --term
│   --record <path>
│   (attach separately via `attach <id>`)
├── ls                           # List active terminals (--json)
├── inspect <id>                 # Terminal details (--json)
├── attach <id>                  # Stream PTY output (Ctrl+C to detach)
├── rm <id>                      # Remove/close terminal (hidden `kill` alias kept)
├── view <id>                    # Print screen content (--format, --json)
├── cursor <id>                  # Print cursor state (--json)
├── size <id>                    # Print dimensions (--json)
├── cell <id> <x> <y>           # Inspect cell (--json)
├── output <id> [bytes]         # Print recent raw PTY output (diagnostics)
├── send <id> <text>            # Send raw text
├── keys <id> <key...>          # Send key events
├── mouse <id> <action> <button> <x> <y>  # Send mouse event
├── paste <id> <text>           # Send paste event
├── focus <id> <gained|lost>    # Send focus event
├── resize <id> <cols> <rows>   # Resize terminal
├── tape <id> <path>            # Play .tape file (--timeout)
├── wait <id> <pattern>         # Wait for regex (--scope, --timeout, --json, exit 1 on timeout)
│   (wait failures include the screen at timeout and recent PTY traffic)
├── verify <id> <golden>        # Verify against golden file (--json, exit 1 on mismatch)
├── snapshot <id> <golden>      # Update/create golden file
├── record <id> <path>          # Start asciinema recording (--stop ends it)
├── marker <id> <label>          # Add recording marker
├── doctor                       # Environment preflight checks (spawns nothing)
├── cast2tape <cast> [-o tape]   # Convert an asciicast recording into a tape
│                                # (input events -> Type/Key/Ctrl/Raw, resizes -> Resize,
│                                #  markers -> comments, WaitStable between clusters)
└── run <tape>                   # Daemon-free: play tape, verify, exit 0/1
    --golden / --update / --cols / --rows / --scrollback / --timeout
    --record / --env / --json (per-tape dir summary) / --strict (no Sleep)
    --term / --jobs N (parallel tapes; N>1 forbids --record)
```

## Code Organization

```
cmd/tabdance-test/
  main.go                    # Entrypoint: --daemon flag, --version (via internal/version)
  cmd/
    root.go                  # Root command var + carapace integration
    mcp.go                   # `mcp` subcommand (init-style self-registration)
    exec.go                  # `exec` subcommand
    ls.go                    # `ls` subcommand
    inspect.go               # `inspect` subcommand
    attach.go                # `attach` subcommand
    rm.go                    # `rm` subcommand (hidden `kill` alias declared here)
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
    record.go                # `record` subcommand (--stop flag)
    marker.go                # `marker` subcommand
    run.go                   # `run` subcommand (daemon-free, in-process runner)
    output.go                # `output` subcommand (recent PTY traffic)
    doctor.go                # `doctor` subcommand (environment preflight)
    flags.go                 # Shared --format/--trim/--unwrap output flag helper
    client.go                # newClient() and parseKeyValue() helpers

pkg/
  actions/
    completion.go             # Public carapace completion actions (ActionTerminalIDs, ActionKeyNames)

internal/
  service/
    service.go               # Transport-neutral operations shared by daemon and MCP
    terminal.go              # Exec/List/Inspect/Delete (with exit-state DTOs)
    input.go                 # SendText/Keys/Mouse/Focus/Paste, Resize
    inspect.go               # View/Cursor/Size/Cell/RecentOutput
    tapeops.go               # PlayTape, Wait (rich timeout context), Verify/UpdateGolden
    run_tape.go              # RunTape: shared operation behind CLI `run` and MCP run_tape (bash OSC integration, strict, exit status)
    recording.go             # Start/Stop recording, markers
  daemon/
    protocol.go              # JSON-RPC protocol types (Request, Response, GetBoolDefault helpers)
    daemon.go                # Daemon: socket listener, accept loop, idle timeout, graceful signal shutdown
    lock.go                  # Instance flock guard (daemon.pid.lock) preventing daemon start races
    handler_terminal.go      # Daemon handlers: exec, ls, inspect, rm
    handler_input.go         # Daemon handlers: send, keys, mouse, paste, focus
    handler_inspect.go       # Daemon handlers: view, cursor, size, cell
    handler_tape.go          # Daemon handlers: resize(input.go), tape, wait
    handler_golden.go        # Daemon handlers: verify, snapshot
    handler_recording.go     # Daemon handlers: record, stop, marker, attach, output
  client/
    client.go                # Client struct, connect, Call
    daemon.go                # Auto-start logic: EnsureDaemon, IsDaemonRunning
    attach.go                # Streaming attach client
    terminal.go              # Terminal methods: Exec, Ls, Inspect, Rm
    inspect.go               # Inspect methods: View, Cursor, Size, Cell (reuse inspect.*Result DTOs)
    input.go                 # Input methods: Send, Keys, Mouse, Paste, Focus, Resize
    tape.go                  # Tape/wait methods: Tape, Wait (reuses wait.WaitForResult)
    golden.go                # Golden methods: Verify, Snapshot (reuses golden.VerifyResult)
    recording.go             # Recording methods: Record, Stop (via `record --stop`), Marker
  server/
    server.go                # Server struct, New(version), Run
    tools.go                 # Tool registration (registerTools)
    handler_terminal.go      # MCP handlers: create, list, close, reset, resize
    handler_input.go         # MCP handlers: send_keys, send_text, send_mouse, send_focus, send_paste
    handler_inspect.go       # MCP handlers: get_view, get_cell, get_cursor, get_size
    handler_recording.go     # MCP handlers: start_recording, stop_recording, add_marker
    handler_tape.go          # MCP handlers: play_tape
    handler_run.go           # MCP handler: run_tape (play + golden verify, strict option, exit info)
    handler_wait.go          # MCP handler: wait_for (failures include screen + recent output)
    handler_golden.go        # MCP handlers: verify_output (errors on mismatch), update_golden
    server_test.go           # Integration tests
  terminal/
    manager.go               # TerminalManager: Create/Get/List/Delete/CloseAll, collision-free random IDs
    managed.go               # ManagedTerminal: terminal + PTY lifecycle, encoders, recorder, mutex
    subscriber.go            # outputSubscriber: broadcast channel for PTY output
    manager_test.go          # Uniqueness/list/delete/name-list tests
    lifecycle_test.go        # Exit-state capture, history ring, group teardown
  input/
    keys.go                  # SendKeys: ParseKey/ParseMods, Kitty keyboard press+release
    mouse.go                 # SendMouse: ParseMouseButton, scroll_up/scroll_down via button 4/5
    focus.go                 # SendFocus: ParseFocusEvent + FocusEncode
    paste.go                 # SendPaste: PasteEncode with bracketed-paste mode
  inspect/
    view.go                  # GetView: Formatter (plain/vt/html/svg/json, trim, unwrap) + cursor + active screen
    style.go                 # cursorStyleString: RenderState-based visual cursor style (shared helper)
    cell.go                  # GetCell: GridRef (codepoint, graphemes, wide, bold, italic, fg/bg, hyperlink)
    cursor.go                # GetCursor: x, y, visible, pending_wrap, style
    size.go                  # GetSize: cols, rows, width_px, height_px
  recording/
    asciicast.go             # AsciicastRecorder: asciicast v2 .cast files (output/input/marker/resize)
  tape/
    executor.go              # TapeExecutor struct, Execute, executeCommand, executeSet
    execute.go               # Tape execution: executeType, sendKeys, executeCtrl/Alt/Shift, executeSleep, executeWait
    extensions.go            # Extension verbs: ExpectExit, WaitExit, WaitStable, Expect; CheckStrict
    parse.go                 # ParseSettings, ShellToCommand, parseCount, parseDuration, parseTypingSpeed, parse*Combo
    executor_test.go         # Tests
  golden/
    verify.go                # VerifyOutput/UpdateGolden: go-difflib unified diff
  wait/
    wait.go                  # WaitFor: regex polling at 50ms, case-insensitive screen/line scope
    wait_test.go             # Scope normalization tests

internal/version/
  version.go                 # Release version var (ldflags-injected), --version fallback

skills/                       # Crush skill references (ghostty, asciinema, vhs)
```

## MCP Tool Summary

22 tools registered in `tools.go:registerTools()`:

|| Tool | Input | Output |
||------|-------|--------|
|| `create_terminal` | cols, rows, scrollback, command, env, working_directory | terminal_id, cols, rows |
|| `list_terminals` | — | terminals[] (id, cols, rows, command, recording, exited, exit_code) |
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
| `get_output` | terminal_id, bytes | recent raw PTY output (control bytes rendered) |
| `run_tape` | tape_path, timeout, golden_path, update, strict | match JSON (+exit_code/exit_signal when the child exited); tool error with diff on mismatch |
|| `verify_output` | terminal_id, golden_path, format, trim, unwrap | JSON match result; tool error with diff on mismatch |
|| `update_golden` | terminal_id, golden_path, format, trim, unwrap | — |
|| `wait_for` | terminal_id, pattern, scope, timeout, format, trim, unwrap | found, elapsed_ms; tool error on timeout |

## Conventions & Patterns

- **Per-terminal mutex**: Each `ManagedTerminal` has its own `sync.Mutex`, exposed via `Lock()`/`Unlock()`. MCP handlers and daemon handlers lock for the duration of their operation.
- **io.Writer on ManagedTerminal**: `Write()` delegates to `WritePty()` so input package can use `io.Writer` interface.
- **Encoder sync after resize**: Always call `SetOptFromTerminal` after resize to keep KeyEncoder/MouseEncoder in sync with terminal mode state.
- **Kitty keyboard protocol**: Send both press and release events when `KittyKeyboardFlags() != KittyKeyDisabled`.
- **VHS parser reuse**: Use `charmbracelet/vhs/lexer` + `vhs/parser` directly. API: `parser.New(lexer)` + `p.Parse()` (check `p.Errors()` separately).
- **TapeExecutor Terminal interface**: Decoupled from `ManagedTerminal` via interface with `Lock/Unlock/WritePty/GetTerminal/GetKeyEncoder/GetMouseEncoder/Resize/AddMarker`.
- **Golden file update mode**: Separate `update_golden` tool (not a flag on `verify_output`) to make accidental updates harder.
- **`Set Shell` only works in `run` mode**: The `run` command pre-parses tape settings via `ParseSettings()` before creating the terminal, so `Set Shell` determines which shell binary runs. For the `tape` MCP tool and CLI `tape` command, the terminal already has a running child process, so `Set Shell` is a no-op at execution time.
- **Key-name normalization**: Use `input.NormalizeKey()` (exported) wherever key names reach `libghostty.ParseKey`. Tape combo parsers import it too; do not create local copies.
- **Scroll mouse actions**: `scroll_up`/`scroll_down` translated to button 4/5 press+release sequences.
- **All go-libghostty getters return `(value, error)`**: e.g., `CursorX() (uint16, error)`, `Cols() (uint16, error)`.
- **mcp-go CallToolRequest**: Use `req.GetString()`, `req.GetInt()`, `req.GetBool()`, `req.RequireString()`, `req.RequireInt()` — NOT `req.Params.Arguments["key"]`.
- **Output subscriber pattern**: `ManagedTerminal` broadcasts PTY output to subscribers via channels. `Subscribe()` returns an `outputSubscriber` with a `Channel()` method. Used by `attach` command to stream output without competing with `ptyReadLoop`.
- **Daemon protocol**: JSON-RPC over Unix socket. Each request is one JSON line, each response is one JSON line. `attach` is special: after the initial request, the daemon streams raw PTY output bytes to the client connection.
- **CLI output format**: Human-readable by default, `--json` flag for machine-parseable output.
- **Init-style subcommand registration**: Each subcommand file declares a package-level `*cobra.Command` var and registers it via `rootCmd.AddCommand()` in an `init()` function. No central list of subcommands in root.go.
- **Carapace completion actions**: Defined in `pkg/actions/` (`ActionTerminalIDs`, `ActionKeyNames`). Command files import and use them directly.
- **Service layer is the only place behavior lives**: `internal/service` owns locking, defaults and error context; daemon and MCP handlers only parse/serialize. Add new operations to the service first, then expose them.
- **`run` command is daemon-free**: Creates terminal in-process, plays tape, verifies, exits. No socket needed. Uses `ParseSettings()` to read `Set Shell`/`Set Width`/`Set Height`/`Set WaitTimeout`/`Set TypingSpeed` from the tape before creating the terminal. Defaults to `sh` if no `Set Shell` is specified.
- **Shared result types**: The client returns `inspect.*Result`, `golden.VerifyResult`, `wait.WaitForResult`, and `terminal.TerminalInfo` directly — daemon, MCP and CLI layers all serialize the same DTOs. Do not re-declare client-local structs.
- **Tri-state bool params**: Daemon requests use `req.GetBoolDefault(key, def)` when absence must fall back to a non-false default (e.g. `trim`/`unwrap`).
- **Wait scope normalization**: `wait.WaitFor` accepts `screen`/`line` case-insensitively (including the `+Line`/`+line` forms); tape `Wait` commands go through the same implementation, so there is exactly one poll loop.
- **Output history ring buffer**: Every `ManagedTerminal` keeps the last 32KB of child output (`RecentOutput(n)`) plus `LastOutputTime()`. Wait/verify failures and the `output` CLI command / `get_output` MCP tool consume it; never read `Ptmx` for history.
- **Child exit state is always recorded**: The reaper goroutine in `startPTY` calls `recordExit`; `ExitState()` returns (exited, exitCode, signal). Tape verbs `ExpectExit`/`WaitExit` and the inspect/list surfaces are built on it.
- **Bash integration is auto-injected by `run`**: tapes with `Set Shell "bash"` get PROMPT_COMMAND/PS0 emitting OSC 133 D/A/C markers (env-only, no rc files, invisible to the rendered screen); `--no-prompt-integration` disables it. Daemon-backed `tape` runs are untouched (existing child process).
- **Determinism guards**: `run --strict` and the `run_tape strict` parameter reject `Sleep` verbs via `tape.CheckStrict`.
- **Failure diagnostics carry context**: wait timeouts and Expect failures include the rendered screen (`--- screen ---`) and recent PTY traffic (`wait.Repr`), not just the pattern.

## Tape Extension Verbs

Lines at the start of a tape line may use seven verbs that tabdance-test executes inline (not part of VHS): `ExpectExit <code>[@dur]`, `WaitExit[@dur]`, `WaitStable[@dur]` (quiet window; the duration doubles as the settle window), `Expect /re/ [+Screen|+Line]` (immediate assertion, no wait), `WaitPrompt[@dur]` / `WaitCommand[<code>][@dur]` (OSC 133 semantic markers; `WaitCommand` also asserts the shell-reported exit status; counting semantics: each verb snapshots the marker counter at execution, so exactly one sync point per command — do not chain both verbs on one command), and `Resize <cols> <rows>` (used by the cast2tape converter; applies at the point the line occurs). `ParseSettings` strips them before the VHS parser; unknown first-word lines still produce VHS parse errors.

## Gotchas

- **libghostty-vt is built from git submodule**: Ghostty source is at `deps/ghostty` (tracking Ghostty `main`). Run `make build-lib` to compile, or any `make build-*` target which depends on it.
- **Zig version**: Ghostty `main` requires Zig 0.17.0 or newer (older Ghostty pins used 0.15.2). Use the `ZIG` make variable to point at the right binary.
- **Build command**: `make build-static` (builds libghostty-vt from submodule + compiles Go). Don't use bare `go build`.
- **`KeyUnknown` doesn't exist**: Use `KeyUnidentified` (value 0) as the zero/unknown sentinel for `libghostty.Key`.
- **`send_keys` accepts both PascalCase and snake_case key names**: `libghostty.ParseKey` only accepts snake_case (e.g. `tab`, `arrow_down`), but `SendKeys` normalizes W3C PascalCase names (e.g. `Tab`, `ArrowDown`) to snake_case automatically via `input.NormalizeKey()`. Prefer PascalCase in MCP requests.
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
- **`mcp` subcommand uses `server.New(version)`**: Not `daemon.New()`. The MCP server is a stdio JSON-RPC server; the daemon is a Unix socket server. The version string comes from `internal/version`.
- **Daemon socket path**: `~/.local/share/tabdance-test/daemon.sock` (override with `TABDANCE_TEST_SOCK`). PID file at `~/.local/share/tabdance-test/daemon.pid`, plus `daemon.pid.lock` (flock held for the daemon's lifetime, never deleted). Socket dir is 0700, socket file 0600.
- **Daemon idle timeout**: 30s after last terminal closed and no client connected (override with `TABDANCE_TEST_IDLE_TIMEOUT`).
- **The child is a session leader with its own process group**: `pty.StartWithSize` sets `Setsid`, so teardown kills the group with negative-PID `syscall.Kill(-pgid, ...)` (SIGTERM then SIGKILL); do not signal `cmd.Process` directly, TUI grandchildren would survive.
- **Wait anchors must match OUTPUT, not the input echo**: typed text appears on screen before Enter is processed, so `Wait+Screen /echo foo/` can fire while Enter is still in flight and race the command's output (first/cold run loses, warm runs win). Anchor on an output-only line, or `WaitStable @250ms` after Enter before verifying.
- **VHS `Wait` syntax is picky**: `Wait+Screen /re/` and `Wait+Line /re/` are the forms with a pattern; bare `Wait` has NO pattern (instant match). Options like `@5s` use `@` (parsed via `parseTypingSpeed`, not `time.ParseDuration`).
- **Daemon sockets are per-project**: socket path hashes the git repo root (cwd fallback); `TABDANCE_TEST_SOCK` overrides the entire resolved path. Separate projects get separate daemons; nested dirs of one repo share theirs.
- **Daemon shutdown is signal-driven** SIGINT/SIGTERM remove the socket and pid files and close all terminals. The flock lock file is intentionally kept on disk (deleting it would race lock waiters).
- **Child processes are reaped exactly once**: `startPTY` spawns a goroutine calling `cmd.Wait()`; do not add another wait — it would degrade into double-wait races. Testing shells must not rely on `SIGTERM` semantics other than the existing SIGTERM-then-SIGKILL escalation in `Close`.

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
make test        # Build libghostty-vt (skipped if cached) + go test ./...
make test-race   # Same with the race detector
make vet         # go vet
make fmt         # gofmt check + fix
```

Unit tests cover: terminal manager (unique IDs, delete semantics, name-list dedupe), daemon instance lock and bool-parameter defaults, key-name normalization, wait-scope normalization, and key combo parsing. Integration tests in `internal/server/server_test.go` cover: create/close, list, get_view, get_cursor, get_size, resize, invalid terminal ID, and golden file verification for nonexistent files. CI (`.github/workflows/build.yml`) runs the goreleaser snapshot build, a gofmt lint job, and `make test` + `make vet`.

## Repository Layout Notes

- `README.md` — user-facing overview, quickstart, and capability matrix.
- `AGENTS.md` — this file: contributor/agent reference.
- `docs/plan.md`, `docs/plan-cli.md` — historical design scratch notes kept for context.
- `skills/` — Crush skill references (ghostty, asciinema, vhs, tabdance-test).

## Cross-Project Skills

- **ghostty** skill — go-libghostty API details (Terminal, KeyEncoder, MouseEncoder, Formatter, GridRef, etc.)
- **asciinema** skill — asciicast v2 format specification
- **vhs** skill — VHS tape DSL and parsing
- **bubbletea** skill — Bubble Tea framework internals
- **tabdance** skill — tabdance application internals
- **carapace** skill — carapace shell completion integration
