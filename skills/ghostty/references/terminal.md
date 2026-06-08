# Terminal — Headless VT Terminal Lifecycle

In-depth reference for the `Terminal` type in `go.mitchellh.com/libghostty`: creation, VT data feeding, resizing, mode control, viewport scrolling, grid references, and lifecycle management.

## Overview

`Terminal` is the central type in go-libghostty. It wraps a `GhosttyTerminal` C handle and provides a **headless, programmable VT terminal** — no GUI, no PTY, no process. You create a terminal, feed it VT-encoded bytes, and inspect its state.

```go
import "go.mitchellh.com/libghostty"

term, err := libghostty.NewTerminal(libghostty.WithSize(80, 24))
if err != nil { log.Fatal(err) }
defer term.Close()

term.VTWrite([]byte("Hello, \033[1;32mworld\033[0m!\r\n"))
```

## Creating a Terminal

```go
func NewTerminal(opts ...TerminalOption) (*Terminal, error)
```

`WithSize(cols, rows)` is **required** — both must be > 0. All other options are optional.

### TerminalOption Functions

| Option | Signature | Description |
|--------|-----------|-------------|
| `WithSize` | `(cols, rows uint16) TerminalOption` | Set terminal dimensions (required) |
| `WithMaxScrollback` | `(lines uint) TerminalOption` | Max scrollback lines (default 0 = none) |
| `WithWritePty` | `(fn WritePtyFn) TerminalOption` | PTY write callback |
| `WithBell` | `(fn BellFn) TerminalOption` | Bell callback |
| `WithTitleChanged` | `(fn TitleChangedFn) TerminalOption` | Title change callback |
| `WithEnquiry` | `(fn EnquiryFn) TerminalOption` | ENQ response callback |
| `WithXtversion` | `(fn XtversionFn) TerminalOption` | XTVERSION response callback |
| `WithSizeReport` | `(fn SizeFn) TerminalOption` | XTWINOPS response callback |
| `WithColorScheme` | `(fn ColorSchemeFn) TerminalOption` | Color scheme query callback |
| `WithDeviceAttributes` | `(fn DeviceAttributesFn) TerminalOption` | DA query callback |

### TerminalConfig

```go
type TerminalConfig struct {
    Cols           uint16
    Rows           uint16
    MaxScrollback  uint
    // + internal effect handler fields
}
```

Can be passed directly to `NewTerminal` or built up with functional options.

## Feeding VT Data

### VTWrite

```go
func (t *Terminal) VTWrite(data []byte)
```

Feeds raw VT-encoded bytes through the terminal's parser, updating terminal state. Malformed input is handled gracefully. **Effect callbacks run synchronously** before this call returns; they must not call `VTWrite` on the same terminal.

### Write (io.Writer)

```go
func (t *Terminal) Write(p []byte) (int, error)
```

Implements `io.Writer`. Always consumes all bytes, never returns an error. Delegates to `VTWrite`.

```go
// Terminal implements io.Writer
fmt.Fprintf(term, "\033[1;1HHello at (1,1)")
```

## Resizing

```go
func (t *Terminal) Resize(cols, rows uint16, cellWidthPx, cellHeightPx uint32) error
```

Changes terminal dimensions. Both `cols` and `rows` must be > 0. `cellWidthPx` and `cellHeightPx` specify pixel dimensions of a single cell, used for image protocols and size reports.

```go
err := term.Resize(120, 40, 8, 16)
```

## Reset

```go
func (t *Terminal) Reset()
```

Full terminal reset (RIS). All state is reset to initial configuration (modes, scrollback, scrolling region, screen contents). Dimensions are preserved.

## Mode Control

```go
func (t *Terminal) ModeGet(mode Mode) (bool, error)
func (t *Terminal) ModeSet(mode Mode, value bool) error
```

Get or set any terminal mode (bracketed paste, mouse tracking, cursor keys, etc.). See [data-types.md](data-types.md) for the full `Mode` constants.

```go
// Check if bracketed paste is active
bp, _ := term.ModeGet(libghostty.ModeBracketedPaste)

// Enable mouse tracking
term.ModeSet(libghostty.ModeMouseAnyEvent, true)
```

## Viewport Scrolling

```go
func (t *Terminal) ScrollViewportTop()
func (t *Terminal) ScrollViewportBottom()
func (t *Terminal) ScrollViewportDelta(delta int)
```

- `ScrollViewportTop()` — Scroll to top of scrollback
- `ScrollViewportBottom()` — Scroll to bottom (active area)
- `ScrollViewportDelta(delta)` — Scroll by delta (negative = up, positive = down)

## Grid References

```go
func (t *Terminal) GridRef(point Point) (*GridRef, error)
func (t *Terminal) TrackGridRef(point Point) (*TrackedGridRef, error)
func (t *Terminal) PointFromGridRef(ref *GridRef, tag PointTag) (Point, error)
```

See [grid-reference.md](grid-reference.md) for full details.

## Kitty Graphics

```go
func (t *Terminal) KittyGraphics() (*KittyGraphics, error)
```

Returns borrowed Kitty graphics image storage. Valid until next mutating call. See [kitty-graphics.md](kitty-graphics.md).

## Data Getters

All data getters are on `*Terminal`. See [data-types.md](data-types.md) for return types.

| Method | Return | Description |
|--------|--------|-------------|
| `Cols()` | `uint16` | Width in cells |
| `Rows()` | `uint16` | Height in cells |
| `CursorX()` | `uint16` | Cursor column (0-indexed) |
| `CursorY()` | `uint16` | Cursor row (0-indexed) |
| `CursorPendingWrap()` | `bool` | Pending wrap flag |
| `CursorVisible()` | `bool` | DEC mode 25 |
| `CursorStyle()` | `*Style` | Current SGR style |
| `ActiveScreen()` | `TerminalScreen` | Primary or Alternate |
| `Title()` | `string` | OSC 0/2 title (borrowed) |
| `Pwd()` | `string` | OSC 7 pwd (borrowed) |
| `MouseTracking()` | `bool` | Any mouse tracking active |
| `KittyKeyboardFlags()` | `KittyKeyFlags` | Kitty keyboard flags |
| `Scrollbar()` | `Scrollbar` | Viewport scrollbar state |
| `TotalRows()` | `uint` | Total rows including scrollback |
| `ScrollbackRows()` | `uint` | Scrollback row count |
| `WidthPx()` | `uint32` | Width in pixels |
| `HeightPx()` | `uint32` | Height in pixels |
| `ViewportActive()` | `bool` | Viewport pinned to active area |
| `ColorForeground()` | `*ColorRGB` | Effective foreground |
| `ColorBackground()` | `*ColorRGB` | Effective background |
| `ColorCursor()` | `*ColorRGB` | Effective cursor color |
| `ColorPalette()` | `*Palette` | Current 256-color palette |
| `ColorForegroundDefault()` | `*ColorRGB` | Default fg (ignoring OSC) |
| `ColorBackgroundDefault()` | `*ColorRGB` | Default bg (ignoring OSC) |
| `ColorCursorDefault()` | `*ColorRGB` | Default cursor (ignoring OSC) |
| `ColorPaletteDefault()` | `*Palette` | Default palette |
| `Selection()` | `*Selection` | Active selection (nil if none) |

### Batch Getters

```go
func (t *Terminal) GetMulti(keys []TerminalData, values []unsafe.Pointer) error
```

Read multiple data fields in a single cgo call (optimized for FFI overhead).

## Option Setters

| Method | Description |
|--------|-------------|
| `SetColorForeground(c *ColorRGB)` | Set default fg color |
| `SetColorBackground(c *ColorRGB)` | Set default bg color |
| `SetColorCursor(c *ColorRGB)` | Set default cursor color |
| `SetColorPalette(palette *Palette)` | Set 256-color palette |
| `SetTitle(title string)` | Set terminal title |
| `SetPwd(pwd string)` | Set working directory |
| `SetSelection(sel *Selection)` | Set/clear active selection |
| `SetAPCMaxBytes(limit *uint)` | APC buffer limit |
| `SetAPCMaxBytesKitty(limit *uint)` | Kitty APC buffer limit |
| `SetKittyImageStorageLimit(limit *uint64)` | Kitty image storage limit |
| `SetKittyImageMediumFile(enabled bool)` | Kitty file medium |
| `SetKittyImageMediumTempFile(enabled bool)` | Kitty temp file medium |
| `SetKittyImageMediumSharedMem(enabled bool)` | Kitty shared mem medium |

## Lifecycle

```go
func (t *Terminal) Close()
```

Frees the underlying terminal handle and releases the `cgo.Handle`. After this call, the terminal must not be used. Always call `Close` when done (typically with `defer`).

## Concurrency

- **Not safe for concurrent use**. Serialize all calls that touch a terminal.
- Effect callbacks run **synchronously during VTWrite** — they must not call VTWrite on the same terminal.
- Borrowed views (GridRef, KittyGraphics, Selection, raw pixel slices) are only valid until the next mutating call.
- Copied values (Cell, Row, Style, ColorRGB, Palette) are regular Go values and may be retained freely.

## TerminalScreen

```go
type TerminalScreen int

const (
    ScreenPrimary   TerminalScreen = C.GHOSTTY_TERMINAL_SCREEN_PRIMARY
    ScreenAlternate TerminalScreen = C.GHOSTTY_TERMINAL_SCREEN_ALTERNATE
)
```

## Scrollbar

```go
type Scrollbar struct {
    Total  uint64  // Total scrollable rows
    Offset uint64  // Viewport offset
    Len    uint64  // Visible rows
}
```

## Typical Headless Testing Pattern

```go
func TestMyTUI(t *testing.T) {
    term, err := libghostty.NewTerminal(
        libghostty.WithSize(80, 24),
        libghostty.WithWritePty(func(_ *libghostty.Terminal, data []byte) {
            // Capture PTY responses (query replies)
        }),
    )
    if err != nil {
        t.Fatal(err)
    }
    defer term.Close()

    // Feed output from the application under test
    term.VTWrite(appOutput)

    // Inspect terminal state
    f, _ := libghostty.NewFormatter(term,
        libghostty.WithFormatterFormat(libghostty.FormatterFormatPlain),
        libghostty.WithFormatterTrim(true),
    )
    defer f.Close()

    content, _ := f.FormatString()
    if !strings.Contains(content, "expected text") {
        t.Errorf("unexpected content: %q", content)
    }
}
```
