# Recording & Testing — go-libghostty Testing Patterns

In-depth reference for integration testing with go-libghostty: headless terminal testing, golden file comparison, and asciicast generation. For asciinema format details, use the **asciinema** skill. For VHS tape format details, use the **vhs** skill.

## Headless Terminal Test Pattern

go-libghostty provides a headless VT terminal for fast, in-process integration testing of TUI applications — no display server, no browser, no PTY required.

### Basic Test

```go
func TestMyTUI(t *testing.T) {
    term, err := libghostty.NewTerminal(
        libghostty.WithSize(80, 24),
        libghostty.WithWritePty(func(_ *libghostty.Terminal, data []byte) {
            ptyResponses <- data
        }),
    )
    if err != nil { t.Fatal(err) }
    defer term.Close()

    term.VTWrite(appOutput)

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

### Simulating Keyboard Input

```go
enc, _ := libghostty.NewKeyEncoder()
defer enc.Close()
enc.SetOptFromTerminal(term)

event := libghostty.NewKeyEvent()
defer event.Close()
event.SetAction(libghostty.KeyActionPress)
event.SetKey(libghostty.KeyEnter)
seq, _ := enc.Encode(event)
```

### Simulating Mouse Input

```go
enc, _ := libghostty.NewMouseEncoder()
defer enc.Close()
enc.SetOptFromTerminal(term)
enc.SetOptSize(libghostty.MouseEncoderSize{
    ScreenWidth: 640, ScreenHeight: 384,
    CellWidth: 8, CellHeight: 16,
})

event := libghostty.NewMouseEvent()
defer event.Close()
event.SetAction(libghostty.MouseActionPress)
event.SetButton(libghostty.MouseButtonLeft)
event.SetPosition(libghostty.MousePosition{X: 10, Y: 5})
seq, _ := enc.Encode(event)
```

### Full Input Pipeline

```go
// 1. Create terminal with PTY response capture
var ptyOut bytes.Buffer
term, _ := libghostty.NewTerminal(
    libghostty.WithSize(80, 24),
    libghostty.WithWritePty(func(_ *libghostty.Terminal, data []byte) {
        ptyOut.Write(data)
    }),
)
defer term.Close()

// 2. Create key encoder synced to terminal state
enc, _ := libghostty.NewKeyEncoder()
defer enc.Close()
enc.SetOptFromTerminal(term)

// 3. Feed application output
term.VTWrite(appOutput)

// 4. Encode and send keyboard input
event := libghostty.NewKeyEvent()
defer event.Close()
event.SetAction(libghostty.KeyActionPress)
event.SetKey(libghostty.KeyEnter)
seq, _ := enc.Encode(event)
```

## Golden File Testing

### Plain Text Comparison

```go
f, _ := libghostty.NewFormatter(term,
    libghostty.WithFormatterFormat(libghostty.FormatterFormatPlain),
    libghostty.WithFormatterTrim(true),
    libghostty.WithFormatterUnwrap(true),
)
defer f.Close()

got, _ := f.FormatString()
want, _ := os.ReadFile("testdata/golden.txt")
if got != string(want) {
    t.Errorf("golden mismatch")
}
```

### VT Snapshot Comparison

```go
f, _ := libghostty.NewFormatter(term,
    libghostty.WithFormatterFormat(libghostty.FormatterFormatVT),
    libghostty.WithFormatterExtraCursor(true),
    libghostty.WithFormatterExtraStyle(true),
    libghostty.WithFormatterExtraPalette(true),
    libghostty.WithFormatterExtraModes(true),
)
defer f.Close()

vtSnapshot, _ := f.Format()
// vtSnapshot can be replayed into another terminal or compared to golden
```

### HTML Snapshot

```go
f, _ := libghostty.NewFormatter(term,
    libghostty.WithFormatterFormat(libghostty.FormatterFormatHTML),
)
defer f.Close()

html, _ := f.FormatString()
os.WriteFile("snapshot.html", []byte(html), 0644)
```

## Asciicast Generation

For asciicast v3 format details, see the **asciinema** skill → `references/asciicast-v3-format.md`.

### Single-Frame Recording

```go
func writeAsciicast(term *libghostty.Terminal, outputPath string) error {
    f, _ := libghostty.NewFormatter(term,
        libghostty.WithFormatterFormat(libghostty.FormatterFormatVT),
    )
    defer f.Close()

    content, _ := f.Format()

    header := map[string]interface{}{
        "version": 2,
        "width":   int(term.Cols()),
        "height":  int(term.Rows()),
        "env":     map[string]string{"TERM": "xterm-ghostty"},
    }

    event := []interface{}{0.0, "o", string(content)}

    headerJSON, _ := json.Marshal(header)
    eventJSON, _ := json.Marshal(event)

    return os.WriteFile(outputPath, append(
        append(headerJSON, '\n'),
        append(eventJSON, '\n')...,
    ), 0644)
}
```

### Incremental Recording with Timing

```go
type AsciicastRecorder struct {
    file  *os.File
    enc   *json.Encoder
    start time.Time
}

func NewAsciicastRecorder(path string, width, height int) (*AsciicastRecorder, error) {
    f, err := os.Create(path)
    if err != nil { return nil, err }

    r := &AsciicastRecorder{file: f, enc: json.NewEncoder(f), start: time.Now()}

    header := map[string]interface{}{
        "version": 2,
        "width":   width,
        "height":  height,
    }
    if err := r.enc.Encode(header); err != nil {
        f.Close()
        return nil, err
    }
    return r, nil
}

func (r *AsciicastRecorder) WriteOutput(data []byte) error {
    elapsed := time.Since(r.start).Seconds()
    return r.enc.Encode([]interface{}{elapsed, "o", string(data)})
}

func (r *AsciicastRecorder) WriteResize(cols, rows int) error {
    elapsed := time.Since(r.start).Seconds()
    return r.enc.Encode([]interface{}{elapsed, "r", fmt.Sprintf("%dx%d", cols, rows)})
}

func (r *AsciicastRecorder) Close() error { return r.file.Close() }
```

## VHS Tape Integration

For VHS tape DSL and parsing details, see the **vhs** skill.

### Converting VHS Tape Commands to go-libghostty

| VHS Command | go-libghostty Equivalent |
|-------------|------------------------|
| `Type "text"` | `KeyEncoder.Encode()` for each character |
| `Enter` | `KeyEncoder.Encode(KeyEvent{Key: KeyEnter})` |
| `Ctrl+C` | `KeyEncoder.Encode(KeyEvent{Key: KeyC, Mods: ModCtrl})` |
| `Sleep 1` | `time.Sleep(1 * time.Second)` |
| `Set Width/Height` | `term.Resize(cols, rows, cellW, cellH)` |
| `Wait /pattern/` | Poll `Formatter.FormatString()` until regex matches |
| `Screenshot` | `RenderState.Update(term)` + custom renderer |

### Parsing VHS Tapes for Headless Execution

```go
import (
    "github.com/charmbracelet/vhs/lexer"
    "github.com/charmbracelet/vhs/parser"
    "github.com/charmbracelet/vhs/token"
)

l := lexer.New(tapeContent)
p := parser.New()
commands, _ := p.Parse(l)

for _, cmd := range commands {
    switch cmd.Type {
    case token.TYPE:
        // Encode each character via KeyEncoder
    case token.ENTER:
        // Encode Enter key
    case token.SLEEP:
        // time.Sleep
    case token.SET:
        // Apply settings (resize, etc.)
    }
}
```

## Comparison: go-libghostty vs VHS

| Feature | go-libghostty | VHS |
|---------|---------------|-----|
| Language | Go library | CLI tool |
| Terminal | Headless VT (no PTY) | Real PTY via ttyd |
| Input | Programmatic (KeyEncoder) | DSL (Type, Ctrl, etc.) |
| Output | Formatter (plain/VT/HTML) | ttyd canvas → ffmpeg |
| Speed | Microseconds | Seconds (browser startup) |
| CI-friendly | Yes (no display needed) | Requires Chromium + ttyd |
| Golden files | Plain text comparison | `.ascii` output |
| Recording | Custom (asciicast generation) | Built-in (GIF/MP4/WebM) |
| Composability | Full Go API | CLI + parser packages only |
