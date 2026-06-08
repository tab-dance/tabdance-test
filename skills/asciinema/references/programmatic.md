# Programmatic — Generating and Using Recordings

In-depth reference for generating asciicast files programmatically, embedding recordings, and using asciinema in testing pipelines.

## Generating .cast Files in Go

### Basic Generation

```go
func writeAsciicast(outputPath string, width, height int, events []Event) error {
    f, err := os.Create(outputPath)
    if err != nil { return err }
    defer f.Close()

    enc := json.NewEncoder(f)

    // Write header
    header := map[string]interface{}{
        "version": 2,
        "width":   width,
        "height":  height,
        "env": map[string]string{
            "TERM":  "xterm-256color",
            "SHELL": "/bin/bash",
        },
    }
    if err := enc.Encode(header); err != nil { return err }

    // Write events
    for _, e := range events {
        event := []interface{}{e.Time, e.Code, e.Data}
        if err := enc.Encode(event); err != nil { return err }
    }

    return nil
}

type Event struct {
    Time float64
    Code string // "o", "i", "m", "r"
    Data string
}
```

### From a go-libghostty Terminal

```go
func recordTerminal(term *libghostty.Terminal, outputPath string) error {
    f, _ := libghostty.NewFormatter(term,
        libghostty.WithFormatterFormat(libghostty.FormatterFormatVT),
    )
    defer f.Close()

    content, _ := f.Format()

    header := map[string]interface{}{
        "version": 2,
        "width":   int(term.Cols()),
        "height":  int(term.Rows()),
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
    file    *os.File
    enc     *json.Encoder
    start   time.Time
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

func (r *AsciicastRecorder) WriteInput(data []byte) error {
    elapsed := time.Since(r.start).Seconds()
    return r.enc.Encode([]interface{}{elapsed, "i", string(data)})
}

func (r *AsciicastRecorder) WriteMarker(label string) error {
    elapsed := time.Since(r.start).Seconds()
    return r.enc.Encode([]interface{}{elapsed, "m", label})
}

func (r *AsciicastRecorder) WriteResize(cols, rows int) error {
    elapsed := time.Since(r.start).Seconds()
    return r.enc.Encode([]interface{}{elapsed, "r", fmt.Sprintf("%dx%d", cols, rows)})
}

func (r *AsciicastRecorder) Close() error {
    return r.file.Close()
}
```

## Parsing .cast Files

```go
func ParseAsciicast(path string) (header map[string]interface{}, events []Event, err error) {
    f, err := os.Open(path)
    if err != nil { return nil, nil, err }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    first := true

    for scanner.Scan() {
        line := scanner.Bytes()

        if first {
            first = false
            if err := json.Unmarshal(line, &header); err != nil {
                return nil, nil, err
            }
            continue
        }

        var event []json.RawMessage
        if err := json.Unmarshal(line, &event); err != nil {
            continue // skip malformed lines
        }

        if len(event) != 3 { continue }

        var timeVal float64
        var code, data string
        json.Unmarshal(event[0], &timeVal)
        json.Unmarshal(event[1], &code)
        json.Unmarshal(event[2], &data)

        events = append(events, Event{Time: timeVal, Code: code, Data: data})
    }

    return header, events, nil
}
```

## Embedding in Web Pages

### asciinema-player

```html
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/asciinema-player@3/dist/bundle.css">
<script src="https://cdn.jsdelivr.net/npm/asciinema-player@3/dist/bundle.js"></script>

<div id="player"></div>
<script>
  AsciinemaPlayer.create('/recordings/demo.cast', document.getElementById('player'), {
    cols: 80,
    rows: 24,
    autoPlay: true,
    loop: true,
    speed: 1.5,
    theme: 'monokai',
    idleTimeLimit: 2,
  });
</script>
```

### Player Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `cols` | int | From recording | Terminal columns |
| `rows` | int | From recording | Terminal rows |
| `autoPlay` | bool | false | Auto-start playback |
| `loop` | bool | false | Loop playback |
| `speed` | float | 1 | Playback speed |
| `theme` | string | — | Theme name (monokai, solarized-dark, etc.) |
| `idleTimeLimit` | float | — | Max idle time between frames |
| `terminalFontSize` | string | — | Font size (CSS value) |
| `fit` | bool | true | Fit to container width |
| `controls` | bool | true | Show playback controls |
| `pauseOnMarkers` | bool | false | Pause on marker events |

## Integration Testing

### Replay-Based Testing

Parse a `.cast` file and replay its output events through a headless terminal, then compare final state:

```go
func TestReplayRecording(t *testing.T) {
    header, events, _ := ParseAsciicast("testdata/recording.cast")

    cols := int(header["width"].(float64))
    rows := int(header["height"].(float64))

    term, _ := libghostty.NewTerminal(libghostty.WithSize(uint16(cols), uint16(rows)))
    defer term.Close()

    for _, e := range events {
        switch e.Code {
        case "o":
            term.VTWrite([]byte(e.Data))
        case "r":
            parts := strings.Split(e.Data, "x")
            c, _ := strconv.Atoi(parts[0])
            r, _ := strconv.Atoi(parts[1])
            term.Resize(uint16(c), uint16(r), 8, 16)
        }
    }

    // Verify terminal state
    f, _ := libghostty.NewFormatter(term,
        libghostty.WithFormatterFormat(libghostty.FormatterFormatPlain),
        libghostty.WithFormatterTrim(true),
    )
    defer f.Close()

    content, _ := f.FormatString()
    if !strings.Contains(content, "expected text") {
        t.Errorf("unexpected content after replay")
    }
}
```

### Recording from Test Harness

Capture test output as asciicast for debugging and documentation:

```go
func TestWithRecording(t *testing.T) {
    rec, _ := NewAsciicastRecorder("testdata/test-recording.cast", 80, 24)
    defer rec.Close()

    // ... run test, capture output ...
    rec.WriteOutput(outputData)

    // Recording saved for post-mortem analysis
}
```
