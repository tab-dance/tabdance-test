# Effects — Terminal→Host Callbacks

In-depth reference for effect callbacks in go-libghostty: how the terminal communicates back to the host application during VT processing.

## Overview

Effects are callbacks invoked **synchronously during `VTWrite`** when the terminal parser encounters sequences that require a host response. They are the mechanism by which a headless terminal communicates back to the embedding application.

**Critical rule**: Effect callbacks must not call `VTWrite` on the same terminal and should avoid blocking for long periods.

## Effect Types

### WritePtyFn — PTY Write Response

```go
type WritePtyFn func(t *Terminal, data []byte)
```

Called when the terminal writes data back to the PTY (e.g., query responses). The `data` slice is only valid for the duration of the call.

```go
libghostty.WithWritePty(func(_ *libghostty.Terminal, data []byte) {
    fmt.Printf("PTY response: %q\n", data)
})
```

This is the most important effect — it fires for CSI query responses (DA, XTWINOPS, DECRQM, etc.), OSC responses, and any other terminal→host communication.

### BellFn — Bell Character

```go
type BellFn func(t *Terminal)
```

Called when the terminal receives a BEL character (0x07).

### TitleChangedFn — Title Change

```go
type TitleChangedFn func(t *Terminal)
```

Called when the terminal title changes via OSC 0 or OSC 2. Read the new title with `t.Title()`.

### EnquiryFn — ENQ Response

```go
type EnquiryFn func(t *Terminal) []byte
```

Called when the terminal receives an ENQ character (0x05). Return the response bytes; nil or empty means no response.

```go
libghostty.WithEnquiry(func(_ *libghostty.Terminal) []byte {
    return []byte("my-terminal")
})
```

### XtversionFn — XTVERSION Response

```go
type XtversionFn func(t *Terminal) string
```

Called for XTVERSION queries (CSI > q). Return the version string; empty uses the default "libghostty".

### SizeFn — XTWINOPS Size Report

```go
type SizeFn func(t *Terminal) (SizeReportSize, bool)
```

Called for XTWINOPS size queries (CSI 14/16/18 t). Return the size and `true`, or zero value and `false` to ignore the query.

```go
libghostty.WithSizeReport(func(_ *libghostty.Terminal) (libghostty.SizeReportSize, bool) {
    return libghostty.SizeReportSize{
        WidthPx:  640,
        HeightPx: 384,
    }, true
})
```

### ColorSchemeFn — Color Scheme Query

```go
type ColorSchemeFn func(t *Terminal) (ColorScheme, bool)
```

Called for color scheme queries (CSI ? 996 n). Return the scheme and `true`, or zero value and `false` to ignore.

### DeviceAttributesFn — Device Attributes

```go
type DeviceAttributesFn func(t *Terminal) (DeviceAttributes, bool)
```

Called for DA queries (CSI c / CSI > c / CSI = c). Return the attributes and `true`, or zero value and `false` to ignore.

## Setting Effects

### At Creation (Functional Options)

```go
term, _ := libghostty.NewTerminal(
    libghostty.WithSize(80, 24),
    libghostty.WithWritePty(func(_ *libghostty.Terminal, data []byte) {
        ptyResponses = append(ptyResponses, data)
    }),
    libghostty.WithBell(func(_ *libghostty.Terminal) {
        bellCount++
    }),
)
```

### On a Live Terminal (SetEffect Methods)

All effects can also be set or cleared on a live terminal via `SetEffect*` methods:

```go
term.SetEffectWritePty(func(_ *libghostty.Terminal, data []byte) { ... })
term.SetEffectBell(func(_ *libghostty.Terminal) { ... })
term.SetEffectTitleChanged(func(_ *libghostty.Terminal) { ... })
// etc.
```

Pass `nil` to clear an effect.

## Internal Mechanism

Effects use **C trampolines**: Go functions are exported via `//export`, registered as C function pointers, and dispatch through `cgo.Handle` userdata to recover the `*Terminal`. The `Terminal.handle` field stores the `cgo.Handle` pointing back to the terminal, set as C-side userdata (`GHOSTTY_TERMINAL_OPT_USERDATA`).

The `effectBuf` field holds C-allocated memory for the most recent response returned by an effect trampoline. libghostty copies the data immediately, so a single buffer shared across effects is sufficient.

## Common Patterns

### Capturing All PTY Output

```go
var ptyBuf bytes.Buffer
term, _ := libghostty.NewTerminal(
    libghostty.WithSize(80, 24),
    libghostty.WithWritePty(func(_ *libghostty.Terminal, data []byte) {
        ptyBuf.Write(data)
    }),
)
```

### Bidirectional Terminal Communication

For a headless terminal that simulates a real PTY session:

```go
// Application output → terminal
term.VTWrite(appOutput)

// Terminal queries → application (via WritePty callback)
libghostty.WithWritePty(func(_ *libghostty.Terminal, data []byte) {
    // These are query responses the terminal wants to send back
    // In a real PTY, these would go to the child process's stdin
    responseChan <- data
})
```

### Testing Query Responses

```go
var lastResponse []byte
term, _ := libghostty.NewTerminal(
    libghostty.WithSize(80, 24),
    libghostty.WithWritePty(func(_ *libghostty.Terminal, data []byte) {
        lastResponse = append([]byte{}, data...)
    }),
    libghostty.WithDeviceAttributes(func(_ *libghostty.Terminal) (libghostty.DeviceAttributes, bool) {
        return libghostty.DeviceAttributes{ /* custom DA */ }, true
    }),
)

// Send DA1 query
term.VTWrite([]byte("\x1b[c"))

// lastResponse now contains the DA response
```
