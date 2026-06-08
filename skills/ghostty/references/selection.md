# Selection — Selection Operations and Gesture State Machine

In-depth reference for the selection system in go-libghostty: selection operations, formatting, adjustment, and the pointer gesture state machine.

## Overview

go-libghostty provides a complete selection system including:
- Programmatic selection operations (word, line, output, all)
- Selection formatting (plain text, VT, HTML)
- Selection adjustment and comparison
- A full pointer gesture state machine for interactive selection

## Selection Type

```go
type Selection struct {
    Start     GridRef
    End       GridRef
    Rectangle bool
}
```

A selection snapshot. Not installed as the active selection by default.

## Programmatic Selection Operations

### SelectWord

```go
func (t *Terminal) SelectWord(options SelectWordOptions) (*Selection, error)
```

```go
type SelectWordOptions struct {
    Ref                 *GridRef   // Grid reference for word selection
    BoundaryCodepoints  []uint32   // Custom word boundaries (nil = defaults)
}
```

Derives a word selection snapshot. Returns nil when no selectable word content.

### SelectWordBetween

```go
func (t *Terminal) SelectWordBetween(options SelectWordBetweenOptions) (*Selection, error)
```

```go
type SelectWordBetweenOptions struct {
    Start               *GridRef   // Starting reference (inclusive)
    End                 *GridRef   // Ending reference (inclusive)
    BoundaryCodepoints  []uint32   // Custom word boundaries
}
```

Searches from Start toward End for the nearest word. Returns nil when no selectable word content exists between references.

### SelectLine

```go
func (t *Terminal) SelectLine(options SelectLineOptions) (*Selection, error)
```

```go
type SelectLineOptions struct {
    Ref                      *GridRef  // Grid reference for line selection
    Whitespace               []uint32  // Custom line-trim whitespace codepoints
    SemanticPromptBoundary    bool      // Bound at semantic prompt state changes
}
```

### SelectAll

```go
func (t *Terminal) SelectAll() (*Selection, error)
```

Selection covering all selectable terminal content. Returns nil when no selectable content.

### SelectOutput

```go
func (t *Terminal) SelectOutput(options SelectOutputOptions) (*Selection, error)
```

```go
type SelectOutputOptions struct {
    Ref *GridRef  // Grid reference within command output
}
```

Semantic command-output selection (requires shell integration / OSC 133 markers). Returns nil when reference is not selectable command output.

## Selection Formatting

### SelectionFormat

```go
func (t *Terminal) SelectionFormat(opts ...SelectionFormatOption) ([]byte, error)
func (t *Terminal) SelectionFormatString(opts ...SelectionFormatOption) (string, error)
func (t *Terminal) SelectionFormatBuf(buf []byte, opts ...SelectionFormatOption) (int, error)
```

Formats the terminal's active selection (or a supplied selection via `WithSelection`).

### SelectionFormatOption Functions

| Option | Signature | Description |
|--------|-----------|-------------|
| `WithSelectionFormat` | `(f FormatterFormat) SelectionFormatOption` | Output format (plain/VT/HTML) |
| `WithSelectionUnwrap` | `(unwrap bool) SelectionFormatOption` | Unwrap soft-wrapped lines |
| `WithSelectionTrim` | `(trim bool) SelectionFormatOption` | Trim trailing whitespace |
| `WithSelection` | `(sel *Selection) SelectionFormatOption` | Format this selection instead of active |

## Selection Adjustment

### SelectionAdjust

```go
func (t *Terminal) SelectionAdjust(sel *Selection, adjustment SelectionAdjust) error
```

Mutates a selection snapshot by adjusting the logical end endpoint.

### SelectionAdjust Constants

| Constant | Description |
|----------|-------------|
| `SelectionAdjustLeft` | Move left to previous non-empty cell, wrapping upward |
| `SelectionAdjustRight` | Move right to next non-empty cell, wrapping downward |
| `SelectionAdjustUp` | Move up one row |
| `SelectionAdjustDown` | Move down to next non-blank row |
| `SelectionAdjustHome` | Move to top-left cell |
| `SelectionAdjustEnd` | Move to right edge of last non-blank row |
| `SelectionAdjustPageUp` | Move up by one page |
| `SelectionAdjustPageDown` | Move down by one page |
| `SelectionAdjustBeginningOfLine` | Move to left edge of current line |
| `SelectionAdjustEndOfLine` | Move to right edge of current line |

## Selection Comparison

### SelectionOrder

```go
func (t *Terminal) SelectionOrder(sel *Selection) (SelectionOrder, error)
```

| Constant | Description |
|----------|-------------|
| `SelectionOrderForward` | Start before End (top-left → bottom-right) |
| `SelectionOrderReverse` | End before Start |
| `SelectionOrderMirroredForward` | Rectangular: top-right → bottom-left |
| `SelectionOrderMirroredReverse` | Rectangular: bottom-left → top-right |

### SelectionOrdered

```go
func (t *Terminal) SelectionOrdered(sel *Selection, desired SelectionOrder) (*Selection, error)
```

Returns a fresh selection with endpoints ordered as requested.

### SelectionContains

```go
func (t *Terminal) SelectionContains(sel *Selection, point Point) (bool, error)
```

Reports whether a point is inside a selection.

### SelectionEqual

```go
func (t *Terminal) SelectionEqual(a *Selection, b *Selection) (bool, error)
```

Reports whether two selections are equal.

## SelectionGesture — Pointer Gesture State Machine

The `SelectionGesture` type implements a full state machine for interpreting pointer events (press/release/drag) as terminal selections, supporting single/double/triple click, word/line/output selection, rectangular selection, and autoscroll.

### Lifecycle

```go
func NewSelectionGesture() (*SelectionGesture, error)
func (g *SelectionGesture) Close(t *Terminal)
func (g *SelectionGesture) Reset(t *Terminal)
```

`Close` takes an optional terminal to release tracked references. `Reset` cancels active gesture state.

### Processing Events

```go
func (g *SelectionGesture) Event(t *Terminal, event *SelectionGestureEvent) (*Selection, error)
```

Applies a gesture event. Returns a selection snapshot, or nil when the event updates state but doesn't produce a selection.

### SelectionGestureEvent

```go
func NewSelectionGestureEvent(eventType SelectionGestureEventType) (*SelectionGestureEvent, error)
func (e *SelectionGestureEvent) Close()
```

#### Event Types

| Type | Description |
|------|-------------|
| `SelectionGestureEventTypePress` | Pointer press |
| `SelectionGestureEventTypeRelease` | Pointer release |
| `SelectionGestureEventTypeDrag` | Pointer drag |
| `SelectionGestureEventTypeAutoscrollTick` | Autoscroll tick |
| `SelectionGestureEventTypeDeepPress` | Deep press (force touch) |

#### Event Options

| Method | Description |
|--------|-------------|
| `SetRef(ref *GridRef)` | Grid reference under pointer |
| `SetPosition(pos SurfacePosition)` | Surface-space pointer position |
| `SetRepeatDistance(d float64)` | Max repeat-click distance in pixels |
| `SetTimeNS(ns uint64)` | Monotonic event time |
| `SetRepeatIntervalNS(ns uint64)` | Max interval between repeat clicks |
| `SetWordBoundaryCodepoints([]uint32)` | Custom word boundaries |
| `SetBehaviors(behaviors SelectionGestureBehaviors)` | Behavior table for press |
| `SetRectangle(rectangle bool)` | Rectangular drag selection |
| `SetGeometry(geometry SelectionGestureGeometry)` | Display geometry |
| `SetViewport(viewport PointCoordinate)` | Viewport for autoscroll |

### SelectionGestureBehavior

| Behavior | Description |
|----------|-------------|
| `SelectionGestureBehaviorCell` | Cell-granular drag |
| `SelectionGestureBehaviorWord` | Word selection on press, word-granular drag |
| `SelectionGestureBehaviorLine` | Line selection on press, line-granular drag |
| `SelectionGestureBehaviorOutput` | Semantic output selection |

### SelectionGestureBehaviors

```go
type SelectionGestureBehaviors struct {
    SingleClick  SelectionGestureBehavior
    DoubleClick  SelectionGestureBehavior
    TripleClick  SelectionGestureBehavior
}
```

### Gesture State Queries

| Method | Return | Description |
|--------|--------|-------------|
| `ClickCount(t *Terminal)` | `uint8` | Current click count (0 = inactive) |
| `Dragged(t *Terminal)` | `bool` | Whether gesture dragged |
| `Autoscroll(t *Terminal)` | `SelectionGestureAutoscroll` | Autoscroll direction |
| `Behavior(t *Terminal)` | `SelectionGestureBehavior` | Current behavior |
| `Anchor(t *Terminal)` | `*GridRef` | Left-click anchor (nil if none) |

### Autoscroll

| Constant | Description |
|----------|-------------|
| `SelectionGestureAutoscrollNone` | No autoscroll |
| `SelectionGestureAutoscrollUp` | Scroll viewport up |
| `SelectionGestureAutoscrollDown` | Scroll viewport down |

## Common Patterns

### Word Selection at Position

```go
point := libghostty.Point{Tag: libghostty.PointTagViewport, X: 10, Y: 5}
ref, _ := term.GridRef(point)
sel, _ := term.SelectWord(libghostty.SelectWordOptions{Ref: ref})
if sel != nil {
    text, _ := term.SelectionFormatString(
        libghostty.WithSelectionFormat(libghostty.FormatterFormatPlain),
        libghostty.WithSelection(sel),
    )
    fmt.Println("Selected word:", text)
}
```

### Interactive Selection via Gesture

```go
gesture, _ := libghostty.NewSelectionGesture()
defer gesture.Close(nil)

// Single click at (10, 5)
pressEvent, _ := libghostty.NewSelectionGestureEvent(libghostty.SelectionGestureEventTypePress)
defer pressEvent.Close()
pressEvent.SetRef(ref)
pressEvent.SetPosition(libghostty.SurfacePosition{X: 80, Y: 80})

sel, _ := gesture.Event(term, pressEvent)

// Drag to (20, 5)
dragEvent, _ := libghostty.NewSelectionGestureEvent(libghostty.SelectionGestureEventTypeDrag)
defer dragEvent.Close()
dragRef, _ := term.GridRef(libghostty.Point{Tag: libghostty.PointTagViewport, X: 20, Y: 5})
dragEvent.SetRef(dragRef)

sel, _ = gesture.Event(term, dragEvent)

// Release
releaseEvent, _ := libghostty.NewSelectionGestureEvent(libghostty.SelectionGestureEventTypeRelease)
defer releaseEvent.Close()
sel, _ = gesture.Event(term, releaseEvent)
```
