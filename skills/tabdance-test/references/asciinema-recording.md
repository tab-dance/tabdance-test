# Asciinema Recording — Capturing Test Sessions

Reference for asciinema recording in tt: creating .cast files from headless terminal sessions.

## Overview

tt can record terminal sessions as asciicast v2 files. This is useful for:
- Debugging test failures (replay what happened)
- Generating documentation
- Creating visual test artifacts

For asciicast v2 format details, see the **asciinema** skill.

## Recording Architecture

```
Terminal Session
  → PTY read goroutine
    → term.VTWrite(data)     [updates terminal state]
    → recorder.WriteOutput(data)  [records to .cast file]
  → Key/Mouse injection
    → ptmx.Write(seq)        [sends to child only; input events are NOT recorded]
  → Resize
    → term.Resize()
    → recorder.WriteResize(cols, rows)
```

## AsciicastRecorder

```go
type AsciicastRecorder struct {
    file    *os.File
    enc     *json.Encoder
    start   time.Time
    idleLimit float64
}

func NewAsciicastRecorder(path string, width, height int, idleLimit float64) (*AsciicastRecorder, error) {
    f, err := os.Create(path)
    if err != nil { return nil, err }

    r := &AsciicastRecorder{
        file:       f,
        enc:        json.NewEncoder(f),
        start:      time.Now(),
        idleLimit:  idleLimit,
    }

    header := map[string]interface{}{
        "version":        2,
        "width":          width,
        "height":         height,
        "env":            map[string]string{"TERM": "xterm-ghostty"},
        "idle_time_limit": idleLimit,
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

## Integration with TerminalManager

```go
type ManagedTerminal struct {
    // ...
    Recorder *AsciicastRecorder
}

func (m *ManagedTerminal) StartRecording(path string, idleLimit float64) error {
    rec, err := NewAsciicastRecorder(path,
        int(m.Terminal.Cols()),
        int(m.Terminal.Rows()),
        idleLimit,
    )
    if err != nil { return err }
    m.Recorder = rec
    return nil
}

func (m *ManagedTerminal) StopRecording() error {
    if m.Recorder == nil { return nil }
    err := m.Recorder.Close()
    m.Recorder = nil
    return err
}
```

## Recording in PTY Bridge

The PTY bridge goroutines write to the recorder when active:

```go
go func() {
    buf := make([]byte, 4096)
    for {
        n, err := ptmx.Read(buf)
        if err != nil { break }

        // Update terminal state
        term.VTWrite(buf[:n])

        // Record output if recording
        if m.Recorder != nil {
            m.Recorder.WriteOutput(buf[:n])
        }
    }
}()
```

## MCP Tool Usage

### Start Recording

```json
{
  "terminal_id": "term_abc123",
  "output_path": "recordings/test-session.cast",
  "idle_time_limit": 2.0
}
```

### Add Marker

```json
{
  "terminal_id": "term_abc123",
  "label": "after login"
}
```

### Stop Recording

```json
{
  "terminal_id": "term_abc123"
}
```

## Playback

Recorded .cast files can be played back with:

```sh
asciinema play recordings/test-session.cast
```

Or embedded in documentation using asciinema-player.
