# VHS Playback — Executing Tape Files

Reference for playing VHS tape files against headless terminals in tabdance-test.

## Overview

tabdance-test can execute VHS `.tape` files against a headless terminal, translating tape DSL commands to go-libghostty operations. This enables using existing VHS tapes for headless testing without requiring Chromium, ttyd, or ffmpeg.

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
| `Enter[N]` | Encode `KeyEnter` N times |
| `Backspace[N]` | Encode `KeyBackspace` N times |
| `Tab[N]` | Encode `KeyTab` N times |
| `Space[N]` | Encode `KeySpace` N times |
| `Ctrl+X` | Encode key with `ModCtrl` |
| `Alt+X` | Encode key with `ModAlt` |
| `Shift+X` | Encode key with `ModShift` |
| `Up/Down/Left/Right[N]` | Encode arrow key N times |
| `PageUp/PageDown[N]` | Encode page key N times |
| `Escape` | Encode `KeyEscape` |
| `Delete/Insert` | Encode `KeyDelete`/`KeyInsert` |
| `Sleep T` | `time.Sleep(parseDuration(T))` |
| `Wait /pattern/` | Poll `Formatter.FormatString()` until regex matches |
| `Set Width/Height` | `term.Resize()` + PTY resize |
| `Set TypingSpeed T` | Set per-character delay |
| `Set FontSize/Shell/Theme/Framerate/PlaybackSpeed/WaitTimeout` | No-op (headless) |
| `Screenshot path` | `AddMarker("screenshot: path")` |
| `Hide` | `AddMarker("hide")` |
| `Show` | `AddMarker("show")` |
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

| Feature | VHS CLI | tabdance-test Tape Executor |
|---------|---------|------------------------------|
| Dependencies | Chromium, ttyd, ffmpeg | None |
| Speed | Seconds (browser startup) | Milliseconds (in-process) |
| CI-friendly | Requires display | No display needed |
| Output | GIF/MP4/WebM | Terminal state + asciinema |
| Composability | CLI only | Full MCP integration |
| Wait assertions | Visual only | Content-based regex |
