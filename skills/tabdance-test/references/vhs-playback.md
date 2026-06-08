# VHS Playback — Executing Tape Files

Reference for playing VHS tape files against headless terminals in tabdance-test.

## Overview

tabdance-test can execute VHS `.tape` files against a headless terminal, translating tape DSL commands to go-libghostty operations. This enables using existing VHS tapes for headless testing without requiring Chromium, ttyd, or ffmpeg.

For VHS tape DSL and parsing details, see the **vhs** skill.

## Tape Executor

```go
type TapeExecutor struct {
    terminal *ManagedTerminal
    timeout  time.Duration
}

func NewTapeExecutor(terminal *ManagedTerminal, timeout time.Duration) *TapeExecutor {
    return &TapeExecutor{terminal: terminal, timeout: timeout}
}
```

## Command Translation

| VHS Command | go-libghostty Operation |
|-------------|------------------------|
| `Type "text"` | `KeyEncoder.Encode()` for each character, write to PTY |
| `Enter[N]` | Encode `KeyEnter` N times |
| `Backspace[N]` | Encode `KeyBackspace` N times |
| `Tab[N]` | Encode `KeyTab` N times |
| `Ctrl+X` | Encode `KeyX` with `ModCtrl` |
| `Alt+X` | Encode with `ModAlt` |
| `Up/Down/Left/Right[N]` | Encode arrow key N times |
| `PageUp/PageDown[N]` | Encode page key N times |
| `Escape` | Encode `KeyEscape` |
| `Delete/Insert` | Encode `KeyDelete`/`KeyInsert` |
| `Sleep T` | `time.Sleep(parseDuration(T))` |
| `Wait /pattern/` | Poll `Formatter.FormatString()` until regex matches |
| `Set Width/Height` | `term.Resize()` + PTY resize |
| `Set FontSize` | No-op (headless has no font) |
| `Set Shell` | No-op (shell already running) |
| `Set TypingSpeed T` | Set per-character delay |
| `Screenshot path` | `RenderState` → custom PNG renderer |
| `Hide` | Pause asciinema recording |
| `Show` | Resume asciinema recording |
| `Copy/Paste` | Clipboard operations (if supported) |
| `Env KEY VALUE` | Set env var (before process start only) |
| `Source path` | Parse and execute included tape |
| `ScrollUp/ScrollDown` | `term.ScrollViewportDelta()` |

## Implementation

```go
func (e *TapeExecutor) Execute(tapePath string) error {
    data, err := os.ReadFile(tapePath)
    if err != nil { return err }

    l := lexer.New(string(data))
    p := parser.New()
    commands, err := p.Parse(l)
    if err != nil { return err }

    typingSpeed := 50 * time.Millisecond

    for _, cmd := range commands {
        switch cmd.Type {
        case token.TYPE:
            text := cmd.Args
            for _, r := range text {
                e.terminal.SendKeyRune(r)
                time.Sleep(typingSpeed)
            }

        case token.ENTER:
            count := parseCount(cmd.Args, 1)
            for i := 0; i < count; i++ {
                e.terminal.SendKey(libghostty.KeyEnter, 0)
            }

        case token.CTRL:
            key, mods := parseCtrlCombo(cmd.Args)
            e.terminal.SendKey(key, mods)

        case token.SLEEP:
            d := parseDuration(cmd.Args)
            time.Sleep(d)

        case token.WAIT:
            pattern := cmd.Args
            scope := cmd.Options // "+Screen" or "+Line"
            timeout := 15 * time.Second
            if err := e.waitForPattern(pattern, scope, timeout); err != nil {
                return fmt.Errorf("wait timeout: %w", err)
            }

        case token.SET:
            e.applySetting(cmd.Options, cmd.Args)

        case token.SCREENSHOT:
            e.terminal.TakeScreenshot(cmd.Args)
        }
    }
    return nil
}
```

## Wait Implementation

```go
func (e *TapeExecutor) waitForPattern(pattern, scope string, timeout time.Duration) error {
    re := regexp.MustCompile(pattern)
    deadline := time.Now().Add(timeout)

    for time.Now().Before(deadline) {
        content := e.terminal.GetViewContent("plain", true, true)

        if scope == "+Screen" {
            if re.MatchString(content) {
                return nil
            }
        } else {
            // Line scope: match current line
            lines := strings.Split(content, "\n")
            cursorY := e.terminal.Terminal.CursorY()
            if int(cursorY) < len(lines) && re.MatchString(lines[cursorY]) {
                return nil
            }
        }

        time.Sleep(50 * time.Millisecond)
    }

    return fmt.Errorf("pattern %q not found within %v", pattern, timeout)
}
```

## MCP Tool Usage

```json
{
  "terminal_id": "term_abc123",
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
