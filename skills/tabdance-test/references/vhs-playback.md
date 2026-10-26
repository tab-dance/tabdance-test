# VHS Playback — Executing Tape Files

Reference for playing VHS tape files against headless terminals in tabdance-test.

## Overview

tt can execute VHS `.tape` files against a headless terminal, translating tape DSL commands to go-libghostty operations. This enables using existing VHS tapes for headless testing without requiring Chromium, ttyd, or ffmpeg.

For VHS tape DSL and parsing details, see the **vhs** skill.

## TapeExecutor

The `TapeExecutor` uses a `Terminal` interface to decouple from `ManagedTerminal`:

```go
type Terminal interface {
    Lock()
    Unlock()
    WritePty(data []byte) (int, error)
    GetTerminal() *libghostty.Terminal
    GetKeyEncoder() *libghostty.KeyEncoder
    GetMouseEncoder() *libghostty.MouseEncoder
    Resize(cols, rows uint16) error
    AddMarker(label string) error
}

type TapeExecutor struct {
    terminal    Terminal
    timeout     time.Duration
    typingSpeed time.Duration
}

func NewTapeExecutor(terminal Terminal, timeout time.Duration) *TapeExecutor
```

## Command Translation

| VHS Command | go-libghostty Operation |
|-------------|------------------------|
| `Type "text"` | `KeyEncoder.Encode()` for each character with `SetUTF8`, write to PTY |
| `Type@time "text"` | Same with per-character delay override |
| `Enter[N]` / `Enter@time[N]` | Encode `KeyEnter` N times, with optional delay between repeats |
| `Backspace[N]` / `Backspace@time[N]` | Encode `KeyBackspace` N times, with optional delay between repeats |
| `Tab[N]` / `Tab@time[N]` | Encode `KeyTab` N times, with optional delay between repeats |
| `Space[N]` / `Space@time[N]` | Encode `KeySpace` N times, with optional delay between repeats |
| `Ctrl+X` | Encode key with `ModCtrl` |
| `Alt+X` | Encode key with `ModAlt` |
| `Shift+X` | Encode key with `ModShift` |
| `Up/Down/Left/Right[N]` | Encode arrow key N times, with delay between repeats |
| `PageUp/PageDown[N]` | Encode page key N times, with delay between repeats |
| `Escape` | Encode `KeyEscape` |
| `Delete/Insert` | Encode `KeyDelete`/`KeyInsert` |
| `Sleep T` | `time.Sleep(parseDuration(T))` |
| `Wait /pattern/` | Poll `Formatter.FormatString()` until regex matches |

Extension verbs (executed inline, not VHS): `ExpectExit <code> @5s`, `WaitExit @5s`, `WaitStable @250ms`, `Expect /re/ [+Screen|+Line]`, `WaitPrompt @5s` (OSC 133 prompt marker), `WaitCommand 0 @5s` (OSC 133 command-finished marker, also asserting the shell-reported exit status), `Resize 80 24` (applies at that point in the tape), and `Raw "\x1b"` (writes literal bytes to the PTY; cast2tape emits it for input bytes with no tape verb). `WaitStable` treats a terminal with zero output so far as stable only after a full quiet window, so a leading WaitStable never races shell startup.

Both OSC verbs snapshot the marker counter when they run: one sync point per command. Prefer `WaitCommand` right after `Enter` (it carries the exit status); `WaitPrompt` asserts the prompt is ready for the NEXT command, so don't put both after the same Enter — the second verb's snapshot lands after both markers have arrived and it hangs until the next command produces markers. `run` injects OSC 133 integration for bash tapes automatically.
| `Wait+Screen /pattern/` | Match against full screen content |
| `Wait+Line /pattern/` | Match against cursor's current line |
| `Set Width/Height` | `term.Resize()` + PTY resize |
| `Set TypingSpeed T` | Set per-character delay |
| `Set Shell` | `ShellToCommand()` maps shell name to command with no-RC flags; applied at terminal creation via `ParseSettings()` |
| `Set WaitTimeout` | Updates default timeout for subsequent `Wait` commands |
| `Set FontSize/FontFamily/Theme/Framerate/PlaybackSpeed` | No-op (headless) |
| `Screenshot path` | `AddMarker("screenshot: path")` |
| `Hide` | Pauses capture on the active recording (asciinema mute): events are dropped and the timeline freezes, so the hidden interval vanishes from the cast. No-op without a recording. |
| `Show` | Resumes capture; the hidden wall-clock time is subtracted (asciinema's time_offset), so the cast shows an uninterrupted session. No-op without a recording. |
| `Output/Require` | No-op |
| `Copy/Paste` | No-op |
| `Env KEY VALUE` | No-op (env set before process start only) |
| `Source path` | Parse and execute included tape recursively |
| `ScrollUp/ScrollDown` | `term.ScrollViewportDelta()` |

## Implementation

```go
func (e *TapeExecutor) Execute(tapePath string) error {
    data, err := os.ReadFile(tapePath)
    if err != nil { return err }

    l := lexer.New(string(data))
    p := parser.New(l)
    commands := p.Parse()

    if errs := p.Errors(); len(errs) > 0 {
        return fmt.Errorf("parse tape: %s", errs[0].Msg)
    }

    for _, cmd := range commands {
        if err := e.executeCommand(cmd); err != nil {
            return fmt.Errorf("execute command %s: %w", cmd.Type, err)
        }
    }
    return nil
}
```

Key differences from VHS CLI parser API:
- `parser.New(l)` takes the lexer as argument (not `parser.New()`)
- `p.Parse()` returns `[]Command` with no error (check `p.Errors()` separately)

## Extension Verbs (tt additions)

Lines starting with these four verbs are executed inline by the tt executor and are NOT part of VHS:

| Verb | Meaning |
|------|---------|
| `ExpectExit <code> @5s` | Wait for the child to exit, then assert its exit code (signal exits never match) |
| `WaitExit @5s` | Wait for the child to exit (any status) |
| `WaitStable @250ms` | Wait until no output for the given quiet window (zero-output terminals must stay empty the whole window) |
| `Expect /ready/ [+Screen|+Line]` | Assert immediately that the pattern is already on screen (no waiting) |

Durations default to the tape's `Set WaitTimeout` / `--timeout`. `ParseSettings` strips these lines before the VHS parser so `Set Shell`/size settings keep working with `run`.

## Wait Implementation

The `Wait` command polls terminal content at 50ms intervals:

```go
func (e *TapeExecutor) executeWait(cmd parser.Command) error {
    pattern := cmd.Args
    scope := cmd.Options  // "+Screen" or "+Line"
    timeout := e.timeout

    re := regexp.MustCompile(pattern)
    deadline := time.Now().Add(timeout)

    for time.Now().Before(deadline) {
        f, _ := libghostty.NewFormatter(e.terminal.GetTerminal(),
            libghostty.WithFormatterFormat(libghostty.FormatterFormatPlain),
            libghostty.WithFormatterTrim(true),
            libghostty.WithFormatterUnwrap(true),
        )
        content, _ := f.FormatString()
        f.Close()

        if scope == "+Screen" || scope == "" {
            if re.MatchString(content) { return nil }
        } else if scope == "+Line" {
            cursorY, _ := e.terminal.GetTerminal().CursorY()
            lines := strings.Split(content, "\n")
            if int(cursorY) < len(lines) && re.MatchString(lines[cursorY]) {
                return nil
            }
        }

        time.Sleep(50 * time.Millisecond)
    }

    return fmt.Errorf("wait timeout: pattern %q not found within %v", pattern, timeout)
}
```

## MCP Tool Usage

```json
{
  "terminal_id": "term_1",
  "tape_path": "tests/login-flow.tape",
  "timeout": 30
}
```

## Advantages Over VHS CLI

| Feature | VHS CLI | tt Tape Executor |
|---------|---------|------------------------------|
| Dependencies | Chromium, ttyd, ffmpeg | None |
| Speed | Seconds (browser startup) | Milliseconds (in-process) |
| CI-friendly | Requires display | No display needed |
| Output | GIF/MP4/WebM | Terminal state + asciinema |
| Composability | CLI only | Full MCP integration |
| Wait assertions | Visual only | Content-based regex |
