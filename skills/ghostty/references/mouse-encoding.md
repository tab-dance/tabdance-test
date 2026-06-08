# Mouse Encoding — Mouse Events to Escape Sequences

In-depth reference for `MouseEncoder` in go-libghostty: encoding mouse events into terminal escape sequences.

## Overview

`MouseEncoder` converts `MouseEvent` structs into the byte sequences a terminal application expects as mouse input. It supports X10, UTF-8, SGR, URxvt, and SGR-Pixels mouse protocols.

```go
enc, _ := libghostty.NewMouseEncoder()
defer enc.Close()
enc.SetOptFromTerminal(term)

event := libghostty.NewMouseEvent()
defer event.Close()
event.SetAction(libghostty.MouseActionPress)
event.SetButton(libghostty.MouseButtonLeft)
event.SetPosition(libghostty.MousePosition{X: 10, Y: 5})

seq, _ := enc.Encode(event)
```

## MouseEncoder

```go
type MouseEncoder struct { ptr C.GhosttyMouseEncoder }
```

Not safe for concurrent use.

### Lifecycle

```go
func NewMouseEncoder() (*MouseEncoder, error)
func (enc *MouseEncoder) Close()
```

### Configuration Methods

| Method | Signature | Description |
|--------|-----------|-------------|
| `SetOptTrackingMode` | `(mode MouseTrackingMode)` | Set mouse tracking mode |
| `SetOptFormat` | `(format MouseFormat)` | Set output format |
| `SetOptSize` | `(s MouseEncoderSize)` | Set renderer size context |
| `SetOptAnyButtonPressed` | `(pressed bool)` | Set whether any button is pressed |
| `SetOptTrackLastCell` | `(track bool)` | Enable motion deduplication |
| `SetOptFromTerminal` | `(t *Terminal)` | Read modes/format from terminal |
| `Reset` | `()` | Clear internal state (e.g., last cell) |

### Encode

```go
func (enc *MouseEncoder) Encode(event *MouseEvent) ([]byte, error)
```

Returns the encoded escape sequence. Not all mouse events produce output; returns `(nil, nil)` in that case.

Uses a 128-byte stack buffer first; on `GHOSTTY_OUT_OF_SPACE`, allocates a heap buffer and retries.

## MouseTrackingMode

```go
type MouseTrackingMode int
```

| Constant | Description |
|----------|-------------|
| `MouseTrackingNone` | Disables mouse reporting |
| `MouseTrackingX10` | X10 mouse mode |
| `MouseTrackingNormal` | Normal mouse mode (button press/release) |
| `MouseTrackingButton` | Button-event tracking |
| `MouseTrackingAny` | Any-event tracking (motion) |

## MouseFormat

```go
type MouseFormat int
```

| Constant | Description |
|----------|-------------|
| `MouseFormatX10` | X10 format (legacy, limited coordinates) |
| `MouseFormatUTF8` | UTF-8 extended format |
| `MouseFormatSGR` | SGR format (CSI < … M/m) |
| `MouseFormatURxvt` | URxvt format |
| `MouseFormatSGRPixels` | SGR format with pixel coordinates |

## MouseEncoderSize

```go
type MouseEncoderSize struct {
    ScreenWidth  uint32 // Full screen width in pixels
    ScreenHeight uint32 // Full screen height in pixels
    CellWidth    uint32 // Cell width in pixels (must be non-zero)
    CellHeight   uint32 // Cell height in pixels (must be non-zero)
    PaddingTop    uint32
    PaddingBottom uint32
    PaddingRight  uint32
    PaddingLeft   uint32
}
```

Required for converting surface-space positions into encoded coordinates.

## MouseEvent

```go
type MouseEvent struct { ptr C.GhosttyMouseEvent }
```

### Lifecycle

```go
func NewMouseEvent() (*MouseEvent, error)
func (e *MouseEvent) Close()
```

### Setters

| Method | Signature | Description |
|--------|-----------|-------------|
| `SetAction` | `(action MouseAction)` | Press, release, move, scroll |
| `SetButton` | `(button MouseButton)` | Mouse button |
| `ClearButton` | `()` | Clear button (for motion events) |
| `SetMods` | `(mods Mods)` | Active modifiers |
| `SetPosition` | `(pos MousePosition)` | Mouse position |

### Getters

| Method | Return | Description |
|--------|--------|-------------|
| `Action()` | `MouseAction` | Event action |
| `Button()` | `MouseButton` | Mouse button |
| `Mods()` | `Mods` | Active modifiers |
| `Position()` | `MousePosition` | Mouse position |

## MouseAction

```go
type MouseAction int
```

| Constant | Description |
|----------|-------------|
| `MouseActionPress` | Button press |
| `MouseActionRelease` | Button release |
| `MouseActionMove` | Mouse motion |
| `MouseActionScrollUp` | Scroll up |
| `MouseActionScrollDown` | Scroll down |
| `MouseActionScrollLeft` | Scroll left |
| `MouseActionScrollRight` | Scroll right |

## MouseButton

```go
type MouseButton int
```

| Constant | Description |
|----------|-------------|
| `MouseButtonLeft` | Left button |
| `MouseButtonMiddle` | Middle button |
| `MouseButtonRight` | Right button |
| `MouseButton4` | Button 4 |
| `MouseButton5` | Button 5 |
| `MouseButton6` | Button 6 |
| `MouseButton7` | Button 7 |
| `MouseButton8` | Button 8 |
| `MouseButtonNone` | No button (motion) |

Implements `String()`, `TextMarshaler`, `TextUnmarshaler`. Parse with `ParseMouseButton(s string) (MouseButton, error)`.

## MousePosition

```go
type MousePosition struct {
    X uint32 // Column (cell-space)
    Y uint32 // Row (cell-space)
}
```

## SetOptFromTerminal

```go
func (enc *MouseEncoder) SetOptFromTerminal(t *Terminal)
```

Reads the terminal's current mouse tracking mode and output format. Does **not** modify size or any-button state.

## Common Patterns

### Simulating Mouse Click

```go
enc, _ := libghostty.NewMouseEncoder()
defer enc.Close()
enc.SetOptFromTerminal(term)
enc.SetOptSize(libghostty.MouseEncoderSize{
    ScreenWidth:  640,
    ScreenHeight: 384,
    CellWidth:    8,
    CellHeight:   16,
})

event := libghostty.NewMouseEvent()
defer event.Close()
event.SetAction(libghostty.MouseActionPress)
event.SetButton(libghostty.MouseButtonLeft)
event.SetPosition(libghostty.MousePosition{X: 10, Y: 5})

seq, _ := enc.Encode(event)
// seq contains the SGR mouse escape sequence
```

### Simulating Scroll

```go
event := libghostty.NewMouseEvent()
defer event.Close()
event.SetAction(libghostty.MouseActionScrollUp)
event.SetPosition(libghostty.MousePosition{X: 40, Y: 12})

seq, _ := enc.Encode(event)
```
