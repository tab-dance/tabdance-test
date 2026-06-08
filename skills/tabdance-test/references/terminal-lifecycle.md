# Terminal Lifecycle — Headless Terminal with PTY

Reference for the headless terminal lifecycle in tabdance-test: creation, PTY management, bidirectional communication, and teardown.

## Overview

tabdance-test creates a go-libghostty `Terminal` for each MCP terminal instance. When a `command` is provided, it also spawns a child process via PTY, creating a full bidirectional terminal session.

## Terminal Creation

### Without PTY (Pure Headless)

```go
term, _ := libghostty.NewTerminal(
    libghostty.WithSize(cols, rows),
    libghostty.WithMaxScrollback(scrollback),
)
```

Use for: testing VT output without a running process, replaying recordings, snapshot comparison.

### With PTY (Full Session)

```go
// Create terminal with PTY write callback
var ptyOut bytes.Buffer
term, _ := libghostty.NewTerminal(
    libghostty.WithSize(cols, rows),
    libghostty.WithMaxScrollback(scrollback),
    libghostty.WithWritePty(func(_ *libghostty.Terminal, data []byte) {
        ptyOut.Write(data)
    }),
)

// Create and sync encoders
keyEnc, _ := libghostty.NewKeyEncoder()
keyEnc.SetOptFromTerminal(term)
mouseEnc, _ := libghostty.NewMouseEncoder()
mouseEnc.SetOptFromTerminal(term)

// Spawn child process
cmd := exec.Command("sh", "-c", command)
cmd.Env = append(os.Environ(), envPairs...)
cmd.Dir = workingDirectory

ptmx, _ := pty.StartWithSize(cmd, &pty.Winsize{
    Cols: uint16(cols),
    Rows: uint16(rows),
})

// Bidirectional bridge:
//   child stdout → term.VTWrite()
//   term PTY responses → child stdin
go func() {
    buf := make([]byte, 4096)
    for {
        n, err := ptmx.Read(buf)
        if err != nil { break }
        m.mu.Lock()
        term.VTWrite(buf[:n])
        if m.Recorder != nil {
            m.Recorder.WriteOutput(buf[:n])
        }
        m.mu.Unlock()
    }
}()
go func() {
    buf := make([]byte, 4096)
    for {
        n, err := ptyOut.Read(buf)
        if err != nil { break }
        if ptmx != nil {
            ptmx.Write(buf[:n])
        }
    }
}()
```

## Bidirectional Communication

```
Child Process (via PTY)
  ↕ stdin/stdout
PTY Master (ptmx)
  ↕ read/write
Bridge Goroutines
  ↕
go-libghostty Terminal
  - VTWrite(child_stdout) → updates terminal state
  - WritePty callback → writes to child stdin
```

### Data Flow

1. **Child → Terminal**: Child writes to stdout → PTY master → bridge goroutine reads → `term.VTWrite(data)` → terminal state updates
2. **Terminal → Child**: Terminal generates query responses → `WritePty` callback fires → bridge goroutine writes to PTY master → child reads from stdin
3. **Agent → Terminal**: Agent sends keys → `KeyEncoder.Encode()` → encoded bytes written to PTY master → child receives input

## ManagedTerminal

```go
type ManagedTerminal struct {
    ID           string
    Terminal     *libghostty.Terminal
    KeyEncoder   *libghostty.KeyEncoder
    MouseEncoder *libghostty.MouseEncoder
    Cmd          *exec.Cmd
    Ptmx         *os.File
    PTYOut       *bytes.Buffer
    Recorder     *AsciicastRecorder
    Command      string
    Cols         uint16
    Rows         uint16
}

// Accessor methods (used by TapeExecutor's Terminal interface)
func (m *ManagedTerminal) GetTerminal() *libghostty.Terminal     { return m.Terminal }
func (m *ManagedTerminal) GetKeyEncoder() *libghostty.KeyEncoder { return m.KeyEncoder }
func (m *ManagedTerminal) GetMouseEncoder() *libghostty.MouseEncoder { return m.MouseEncoder }

// Mutex access
func (m *ManagedTerminal) Lock()
func (m *ManagedTerminal) Unlock()

// PTY I/O
func (m *ManagedTerminal) WritePty(data []byte) (int, error)
func (m *ManagedTerminal) Write(data []byte) (int, error)  // io.Writer, delegates to WritePty
```

## Terminal Manager

```go
type TerminalManager struct {
    mu         sync.Mutex
    terminals  map[string]*ManagedTerminal
    counter    int
}

func (m *TerminalManager) Create(cfg CreateConfig) (*ManagedTerminal, error)
func (m *TerminalManager) Get(id string) (*ManagedTerminal, error)
func (m *TerminalManager) List() []TerminalInfo
func (m *TerminalManager) Delete(id string) error
func (m *TerminalManager) CloseAll()
```

Terminal IDs are auto-generated as `term_N` (incrementing counter).

## Resize Handling

When the agent resizes a terminal:

1. `term.Resize(cols, rows, cellWidthPx, cellHeightPx)` — updates go-libghostty terminal
2. `keyEnc.SetOptFromTerminal(term)` + `mouseEnc.SetOptFromTerminal(term)` — re-sync encoders
3. If PTY is active: `pty.Setsize(ptmx, &pty.Winsize{Cols, Rows})` — sends SIGWINCH to child
4. If recording: `recorder.WriteResize(cols, rows)` — records resize event

## Teardown

```go
func (m *ManagedTerminal) Close() {
    // 1. Signal done channel (stops bridge goroutines)
    close(m.done)

    // 2. Stop recording if active
    if m.Recorder != nil {
        m.Recorder.Close()
    }

    // 3. Close encoders
    m.KeyEncoder.Close()
    m.MouseEncoder.Close()

    // 4. Kill child process
    if m.Cmd != nil && m.Cmd.Process != nil {
        m.Cmd.Process.Signal(syscall.SIGTERM)
        time.Sleep(100 * time.Millisecond)
        m.Cmd.Process.Kill()
    }

    // 5. Close PTY
    if m.Ptmx != nil {
        m.Ptmx.Close()
    }

    // 6. Close terminal last
    m.Terminal.Close()
}
```

Server shutdown calls `s.terminals.CloseAll()` via `defer` in `Run()`.

## Concurrency

- Each `ManagedTerminal` has its own `sync.Mutex`, exposed via `Lock()`/`Unlock()` methods
- All terminal access must be serialized via the per-terminal mutex
- The `WritePty` callback runs synchronously during `VTWrite` — must not call `VTWrite` on the same terminal
- The PTY bridge goroutines hold the mutex during `VTWrite` + recorder write
- MCP handlers lock the terminal for the duration of their operation
- Bridge goroutines exit when the `done` channel is closed
