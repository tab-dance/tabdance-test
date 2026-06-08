# Recording & Testing — Asciinema, VHS, and Golden Files

In-depth reference for terminal recording formats (asciinema v2, VHS tapes) and integration testing patterns using go-libghostty.

## Asciicast v2 Format

The asciinema recording format (`.cast` files) is based on **newline-delimited JSON** (JSON Lines).

### Header (First Line)

```json
{
  "version": 2,
  "width": 80,
  "height": 24,
  "timestamp": 1504467315,
  "duration": 123.45,
  "command": "/bin/bash",
  "title": "Demo",
  "idle_time_limit": 2,
  "env": {"SHELL": "/bin/bash", "TERM": "xterm-256color"},
  "theme": {
    "fg": "#d0d0d0",
    "bg": "#212121",
    "palette": "#151515:#ac4142:..."
  }
}
```

**Required**: `version` (must be 2), `width`, `height`.

### Event Stream (Subsequent Lines)

Each line is a 3-element JSON array: `[time, code, data]`

| Code | Type | Description |
|------|------|-------------|
| `"o"` | Output | Data written to terminal (string) |
| `"i"` | Input | Keystrokes from terminal (string) |
| `"m"` | Marker | Breakpoint/label (string, can be empty) |
| `"r"` | Resize | New size as `"{COLS}x{ROWS}"` (string) |

**Timing**: `time` is seconds since start of recording (float).

### Example

```
{"version": 2, "width": 80, "height": 24}
[0.248848, "o", "\u001b[1;31mHello \u001b[32mWorld!\u001b[0m\n"]
[1.001376, "o", "That was ok\rThis is better."]
[1.500000, "m", ""]
[2.143733, "o", "Now... "]
[4.050000, "r", "80x24"]
[6.541828, "o", "Bye!"]
```

### File Properties

| Property | Value |
|----------|-------|
| Extension | `.cast` |
| MIME type | `application/x-asciicast` |

### Design Benefits

- Incremental, real-time writing to disk
- Minimal memory usage (disk-space limited)
- Crash safety (recording not lost on interruption)
- Real-time streaming friendly

## VHS (charmbracelet/vhs) — Tape Format

VHS is a CLI tool that executes `.tape` files to produce terminal GIFs, MP4s, WebMs, or text output for testing.

### Architecture

Three-process pipeline:
1. **ttyd** — Terminal multiplexer hosting a real PTY via xterm.js
2. **go-rod (Chromium)** — Headless Chrome connected to ttyd
3. **ffmpeg** — Composites frames into final output

### Tape DSL Commands

| Command | Syntax | Description |
|---------|--------|-------------|
| `Output` | `Output <path>` | Output file (`.gif`, `.mp4`, `.webm`, `.txt`, `.ascii`) |
| `Require` | `Require <program>` | Fail if program not in `$PATH` |
| `Set` | `Set <Setting> <Value>` | Configure terminal/recording |
| `Type` | `Type[@time] "string"` | Type characters |
| `Enter` | `Enter[count]` | Press Enter |
| `Backspace` | `Backspace[count]` | Press Backspace |
| `Tab` | `Tab[@time] [count]` | Press Tab |
| `Ctrl` | `Ctrl[+Alt][+Shift]+<char>` | Control sequence |
| `Alt` | `Alt+<char>` | Alt modifier |
| `Sleep` | `Sleep <time>` | Wait (captures frames) |
| `Wait` | `Wait[+Screen\|+Line][@timeout] /regex/` | Wait for regex match |
| `Hide` | `Hide` | Stop capturing frames |
| `Show` | `Show` | Resume capturing |
| `Screenshot` | `Screenshot <path.png>` | Capture frame as PNG |
| `Copy` | `Copy "text"` | Copy to clipboard |
| `Paste` | `Paste` | Paste from clipboard |
| `Env` | `Env <KEY> "value"` | Set environment variable |
| `Source` | `Source <path.tape>` | Include another tape |

### Set Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `Shell` | `bash` | Shell to use |
| `FontSize` | `22` | Font size |
| `FontFamily` | JetBrains Mono | Font family |
| `Width` | `1200` | Pixel width |
| `Height` | `600` | Pixel height |
| `TypingSpeed` | `50ms` | Per-character delay |
| `Theme` | Default dark | Theme name or JSON |
| `Framerate` | `50` | Capture framerate |
| `WaitTimeout` | `15s` | Wait command timeout |
| `WaitPattern` | `/>$/` | Default wait pattern |

### VHS for Testing

When output is `.txt` or `.ascii`, VHS saves the terminal buffer content after each command. These golden files can be stored in git and diffed for integration testing.

**VHS is NOT a Go library** — it's a CLI application. The lexer/parser packages are importable:
- `github.com/charmbracelet/vhs/lexer`
- `github.com/charmbracelet/vhs/parser`
- `github.com/charmbracelet/vhs/token`

## Integration Testing with go-libghostty

go-libghostty provides a superior alternative to VHS for Go-based integration testing of TUI applications.

### Headless Terminal Test Pattern

```go
func TestMyTUI(t *testing.T) {
    // 1. Create headless terminal
    term, err := libghostty.NewTerminal(
        libghostty.WithSize(80, 24),
        libghostty.WithWritePty(func(_ *libghostty.Terminal, data []byte) {
            ptyResponses <- data
        }),
    )
    if err != nil { t.Fatal(err) }
    defer term.Close()

    // 2. Feed application output
    term.VTWrite(appOutput)

    // 3. Inspect terminal state
    f, _ := libghostty.NewFormatter(term,
        libghostty.WithFormatterFormat(libghostty.FormatterFormatPlain),
        libghostty.WithFormatterTrim(true),
        libghostty.WithFormatterUnwrap(true),
    )
    defer f.Close()

    got, _ := f.FormatString()
    want, _ := os.ReadFile("testdata/golden.txt")
    if got != string(want) {
        t.Errorf("mismatch:\ngot:  %q\nwant: %q", got, string(want))
    }
}
```

### Simulating Input

```go
// Create key encoder synced to terminal state
enc, _ := libghostty.NewKeyEncoder()
defer enc.Close()
enc.SetOptFromTerminal(term)

// Simulate key press
event := libghostty.NewKeyEvent()
defer event.Close()
event.SetAction(libghostty.KeyActionPress)
event.SetKey(libghostty.KeyEnter)
seq, _ := enc.Encode(event)

// In a real test, seq would be written to the application's stdin
```

### Asciinema Recording Generation

```go
func writeAsciicast(term *libghostty.Terminal, outputPath string) error {
    header := map[string]interface{}{
        "version": 2,
        "width":   int(term.Cols()),
        "height":  int(term.Rows()),
        "env": map[string]string{
            "TERM": "xterm-ghostty",
        },
    }

    // Format terminal content as VT
    f, _ := libghostty.NewFormatter(term,
        libghostty.WithFormatterFormat(libghostty.FormatterFormatVT),
    )
    defer f.Close()

    content, _ := f.Format()

    // Write asciicast file
    headerJSON, _ := json.Marshal(header)
    event := []interface{}{0.0, "o", string(content)}
    eventJSON, _ := json.Marshal(event)

    return os.WriteFile(outputPath, append(
        append(headerJSON, '\n'),
        append(eventJSON, '\n')...,
    ), 0644)
}
```

### VHS Tape Parsing

```go
import (
    "github.com/charmbracelet/vhs/lexer"
    "github.com/charmbracelet/vhs/parser"
)

// Parse a .tape file
l := lexer.New(tapeContent)
p := parser.New()
commands, err := p.Parse(l)
if err != nil { return err }

// commands is a []Command — each has Type, Options, Args
for _, cmd := range commands {
    switch cmd.Type {
    case token.TYPE:
        // Simulate typing
    case token.ENTER:
        // Simulate Enter key
    case token.SLEEP:
        // Wait
    }
}
```

### Golden File Testing Pattern

```go
func TestGoldenFile(t *testing.T) {
    term, _ := libghostty.NewTerminal(libghostty.WithSize(80, 24))
    defer term.Close()

    // Feed known input
    term.VTWrite([]byte("\033[1;1HHello, World!\033[0K"))

    // Extract plain text
    f, _ := libghostty.NewFormatter(term,
        libghostty.WithFormatterFormat(libghostty.FormatterFormatPlain),
        libghostty.WithFormatterTrim(true),
    )
    defer f.Close()

    got, _ := f.FormatString()

    // Update golden: os.WriteFile("testdata/golden.txt", []byte(got), 0644)
    want, _ := os.ReadFile("testdata/golden.txt")
    if got != string(want) {
        t.Errorf("golden mismatch")
    }
}
```

### Screenshot/Visual Testing

For pixel-level visual testing, use the `RenderState` to build a custom renderer that produces images:

```go
rs, _ := libghostty.NewRenderState()
defer rs.Close()
rs.Update(term)

// Iterate rows/cells, build image
// Compare against golden PNG
```

## Comparison: go-libghostty vs VHS

| Feature | go-libghostty | VHS |
|---------|---------------|-----|
| Language | Go library | CLI tool (Go internally) |
| Terminal | Headless VT (no PTY) | Real PTY via ttyd |
| Input | Programmatic (KeyEncoder) | DSL (Type, Ctrl, etc.) |
| Output | Formatter (plain/VT/HTML) | ttyd canvas → ffmpeg |
| Speed | Microseconds | Seconds (browser startup) |
| CI-friendly | Yes (no display needed) | Requires Chromium + ttyd |
| Golden files | Plain text comparison | `.ascii` output |
| Recording | Custom (asciicast generation) | Built-in (GIF/MP4/WebM) |
| Composability | Full Go API | CLI + parser packages only |
