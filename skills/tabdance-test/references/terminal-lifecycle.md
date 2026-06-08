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
    libghostty.WithWritePty(func(_ *libghostty.Terminal, data []byte) {
        ptyOut.Write(data)
    }),
)

// Spawn child process
cmd := exec.Command(command, args...)
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
        term.VTWrite(buf[:n])
    }
}()
go func() {
    buf := make([]byte, 4096)
    for {
        n, err := ptyOut.Read(buf)
        if err != nil { break }
        ptmx.Write(buf[:n])
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

## Terminal Manager

tabdance-test maintains a map of active terminals:

```go
type TerminalManager struct {
    mu         sync.Mutex
    terminals  map[string]*ManagedTerminal
}

type ManagedTerminal struct {
    ID         string
    Terminal   *libghostty.Terminal
    KeyEncoder *libghostty.KeyEncoder
    MouseEncoder *libghostty.MouseEncoder
    Cmd        *exec.Cmd
    Ptmx       *os.File
    PTYOut     *bytes.Buffer  // captures WritePty output
    Recorder   *AsciicastRecorder
}
```

## Resize Handling

When the agent resizes a terminal:

1. `term.Resize(cols, rows, cellWidthPx, cellHeightPx)` — updates go-libghostty terminal
2. If PTY is active: `pty.Setsize(ptmx, &pty.Winsize{Cols, Rows})` — sends SIGWINCH to child
3. `enc.SetOptFromTerminal(term)` — re-sync key encoder to new terminal state

## Teardown

```go
func (m *ManagedTerminal) Close() {
    // 1. Stop recording if active
    if m.Recorder != nil {
        m.Recorder.Close()
    }

    // 2. Close encoders
    m.KeyEncoder.Close()
    m.MouseEncoder.Close()

    // 3. Kill child process
    if m.Cmd != nil && m.Cmd.Process != nil {
        m.Cmd.Process.Signal(syscall.SIGTERM)
        time.Sleep(100 * time.Millisecond)
        m.Cmd.Process.Kill()
    }

    // 4. Close PTY
    if m.Ptmx != nil {
        m.Ptmx.Close()
    }

    // 5. Close terminal last
    m.Terminal.Close()
}
```

## Concurrency

- Each `ManagedTerminal` is accessed by multiple goroutines (PTY read, PTY write, MCP handler)
- All terminal access must be serialized via a per-terminal mutex
- The `WritePty` callback runs synchronously during `VTWrite` — must not call `VTWrite` on the same terminal
- The PTY bridge goroutines must not call terminal methods directly; they write to/read from buffers
