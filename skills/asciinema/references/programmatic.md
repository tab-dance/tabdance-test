# Programmatic — Generating and Using Recordings

In-depth reference for generating asciicast **v3** files programmatically, embedding recordings, and using asciinema in testing pipelines.

## Generating .cast Files in Go

### Basic Generation

In asciicast v3 the header carries the terminal inside a `term` object, and event lines carry **intervals** (seconds since the previous event):

```go
func writeAsciicast(outputPath string, width, height int, events []Event) error {
    f, err := os.Create(outputPath)
    if err != nil { return err }
    defer f.Close()

    enc := json.NewEncoder(f)

    // Write header
    header := map[string]interface{}{
        "version": 3,
        "term": map[string]interface{}{
            "cols": width,
            "rows": height,
            "type": "xterm-256color",
        },
        "env": map[string]string{
            "SHELL": "/bin/bash",
        },
    }
    if err := enc.Encode(header); err != nil { return err }

    // Write events (interval since the previous event)
    for _, e := range events {
        event := []interface{}{e.Interval, e.Code, e.Data}
        if err := enc.Encode(event); err != nil { return err }
    }

    return nil
}

type Event struct {
    Interval float64
    Code     string // "o", "i", "m", "r", "x"
    Data     string
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
        "version": 3,
        "term": map[string]interface{}{
            "cols": int(term.Cols()),
            "rows": int(term.Rows()),
        },
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

Events are `[interval, code, data]`; track the delay since the previous event explicitly:

```go
type AsciicastRecorder struct {
    file      *os.File
    enc       *json.Encoder
    start     time.Time
    mu        sync.Mutex
    elapsed   float64
}

func NewAsciicastRecorder(path string, width, height int) (*AsciicastRecorder, error) {
    f, err := os.Create(path)
    if err != nil { return nil, err }

    r := &AsciicastRecorder{file: f, enc: json.NewEncoder(f), start: time.Now()}

    header := map[string]interface{}{
        "version": 3,
        "term": map[string]interface{}{
            "cols": width,
            "rows": height,
        },
    }
    if err := r.enc.Encode(header); err != nil {
        f.Close()
        return nil, err
    }

    return r, nil
}

// interval returns the seconds since the previously emitted event and
// advances the bookkeeping; callers serialize writers with the mutex.
func (r *AsciicastRecorder) interval() float64 {
    gap := time.Since(r.start).Seconds() - r.elapsed
    r.elapsed += gap
    return gap
}

func (r *AsciicastRecorder) WriteOutput(data []byte) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    return r.enc.Encode([]interface{}{r.interval(), "o", string(data)})
}

func (r *AsciicastRecorder) WriteInput(data []byte) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    return r.enc.Encode([]interface{}{r.interval(), "i", string(data)})
}

func (r *AsciicastRecorder) WriteMarker(label string) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    return r.enc.Encode([]interface{}{r.interval(), "m", label})
}

func (r *AsciicastRecorder) WriteResize(cols, rows int) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    return r.enc.Encode([]interface{}{r.interval(), "r", fmt.Sprintf("%dx%d", cols, rows)})
}

func (r *AsciicastRecorder) Close() error {
    // Trailing no-op event keeps the quiet tail in playback.
    r.mu.Lock()
    _ = r.enc.Encode([]interface{}{r.interval(), "o", ""})
    r.mu.Unlock()
    return r.file.Close()
}
```

## Parsing .cast Files

The first line is the v3 header (`version` must be 3, terminal inside `term`); skip `#` comment lines; accumulate event intervals if you need absolute times:

```go
func ParseAsciicast(path string) (header map[string]interface{}, events []Event, err error) {
    f, err := os.Open(path)
    if err != nil { return nil, nil, err }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    first := true
    abs := 0.0

    for scanner.Scan() {
        line := scanner.Bytes()
        if len(line) == 0 || line[0] == '#' {
            continue
        }

        if first {
            first = false
            if err := json.Unmarshal(line, &header); err != nil {
                return nil, nil, err
            }
            if int(header["version"].(float64)) != 3 {
                return nil, nil, fmt.Errorf("unsupported asciicast version")
            }
            continue
        }

        var event []json.RawMessage
        if err := json.Unmarshal(line, &event); err != nil {
            continue // skip malformed lines
        }

        if len(event) != 3 { continue }

        var interval float64
        var code, data string
        json.Unmarshal(event[0], &interval)
        json.Unmarshal(event[1], &code)
        json.Unmarshal(event[2], &data)

        abs += interval
        events = append(events, Event{Time: abs, Code: code, Data: data})
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

    term := header["term"].(map[string]interface{})
    cols := int(term["cols"].(float64))
    rows := int(term["rows"].(float64))

    termInst, _ := libghostty.NewTerminal(libghostty.WithSize(uint16(cols), uint16(rows)))
    defer termInst.Close()

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
