# Asciinema Recording — Capturing Test Sessions

Reference for asciinema recording in tt: creating .cast files from headless terminal sessions.

## Overview

tt records terminal sessions as **asciicast v3** files (interval-timed JSON Lines; see the **asciinema** skill for the format spec). This is useful for:
- Debugging test failures (replay what happened)
- Generating documentation
- Creating visual test artifacts

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

Produces asciicast **v3**: header with `version: 3` and a `term` object (`cols`/`rows`/`type`), then events whose first element is the **interval since the previous event**:

```json
{"version": 3, "term": {"cols": 80, "rows": 24, "type": "xterm-ghostty"}, "idle_time_limit": 0}
[0.002164768, "o", "\u001b[?2004h"]
[0.10047151, "i", "c"]
```

```go
type AsciicastRecorder struct {
    file      *os.File
    enc       *json.Encoder
    start     time.Time
    idleLimit float64
    elapsed   float64 // cumulative emitted time, trimmed by idleLimit
    mu        sync.Mutex
}

func (r *AsciicastRecorder) intervalLocked() float64 {
    gap := time.Since(r.start).Seconds() - r.elapsed
    if r.idleLimit > 0 && r.elapsed > 0 && gap > r.idleLimit {
        gap = r.idleLimit
    }
    r.elapsed += gap
    return gap
}
```

Event writers emit `[interval, code, data]` for codes `o` (output), `i` (input), `m` (marker label), and `r` (`"COLSxROWS"`):

```go
func (r *AsciicastRecorder) WriteOutput(data []byte) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    return r.enc.Encode([]any{r.intervalLocked(), "o", string(data)})
}

func (r *AsciicastRecorder) WriteInput(data []byte) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    return r.enc.Encode([]any{r.intervalLocked(), "i", string(data)})
}

func (r *AsciicastRecorder) WriteMarker(label string) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    return r.enc.Encode([]any{r.intervalLocked(), "m", label})
}

func (r *AsciicastRecorder) WriteResize(cols, rows int) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    return r.enc.Encode([]any{r.intervalLocked(), "r", fmt.Sprintf("%dx%d", cols, rows)})
}
```

## Idle Handling and the Trailing Idle Event

`Close` appends one no-op output event (`[interval, "o", ""]`) at the true stop time:

```go
func (r *AsciicastRecorder) Close() error {
    r.mu.Lock()
    _ = r.enc.Encode([]any{r.intervalLocked(), "o", ""})
    r.mu.Unlock()
    return r.file.Close()
}
```

Without it, a cast issued during quiet time (e.g. a trailing `Sleep 5s` in a tape)
would end at its last real event and playback would stop early. With `idleLimit > 0`
(daemon/MCP recordings default to 2.0) the interval is clamped instead, so long dead
tails are shortened. `tt run --record` writes `idle_time_limit: 0` (no clamping).

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

Recorded v3 .cast files can be played back with:

```sh
asciinema play recordings/test-session.cast
```

Playback needs asciinema CLI 3.x (or player/server v3.10+); older players only
understand v2 — use `asciinema convert` if you must hand the file to one.
