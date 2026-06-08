# Render State — High-Performance Viewport Rendering

In-depth reference for `RenderState` in go-libghostty: optimized viewport state with dirty tracking for incremental rendering.

## Overview

`RenderState` holds a snapshot of the terminal's visible viewport, optimized for repeated updates. It tracks dirty regions so a renderer can skip unchanged rows. After `Update`, the render state can be read without touching the terminal until the next `Update`.

```go
rs, _ := libghostty.NewRenderState()
defer rs.Close()

// Update from terminal (requires exclusive access during call)
err := rs.Update(term)

// Read render state without touching the terminal
// ... iterate rows, read cells, check dirty flags
```

## RenderState

```go
type RenderState struct { ptr C.GhosttyRenderState }
```

### Lifecycle

```go
func NewRenderState() (*RenderState, error)
func (rs *RenderState) Close()
```

### Update

```go
func (rs *RenderState) Update(t *Terminal) error
```

Updates the render state from a terminal. This consumes terminal/screen dirty state and is the **only** render-state operation that touches the terminal. Hold exclusive access to the terminal while this call is running.

After `Update` returns, the render state can be read **without** touching the terminal until the next `Update`.

## RenderStateDirty

```go
type RenderStateDirty int

const (
    RenderStateDirtyFalse   RenderStateDirty = C.GHOSTTY_RENDER_STATE_DIRTY_FALSE
    RenderStateDirtyPartial RenderStateDirty = C.GHOSTTY_RENDER_STATE_DIRTY_PARTIAL
    RenderStateDirtyFull    RenderStateDirty = C.GHOSTTY_RENDER_STATE_DIRTY_FULL
)
```

| Value | Meaning |
|-------|---------|
| `RenderStateDirtyFalse` | Not dirty; rendering can be skipped |
| `RenderStateDirtyPartial` | Some rows changed; incremental redraw |
| `RenderStateDirtyFull` | Global state changed; full redraw |

## CursorVisualStyle

```go
type CursorVisualStyle int

const (
    CursorVisualStyleBar         CursorVisualStyle = C.GHOSTTY_RENDER_STATE_CURSOR_VISUAL_STYLE_BAR
    CursorVisualStyleBlock        CursorVisualStyle = C.GHOSTTY_RENDER_STATE_CURSOR_VISUAL_STYLE_BLOCK
    CursorVisualStyleUnderline    CursorVisualStyle = C.GHOSTTY_RENDER_STATE_CURSOR_VISUAL_STYLE_UNDERLINE
    CursorVisualStyleBlockHollow CursorVisualStyle = C.GHOSTTY_RENDER_STATE_CURSOR_VISUAL_STYLE_BLOCK_HOLLOW
)
```

| Style | DECSCUSR | Description |
|-------|----------|-------------|
| `Bar` | 5, 6 | Vertical bar cursor |
| `Block` | 1, 2 | Filled block cursor |
| `Underline` | 3, 4 | Underline cursor |
| `BlockHollow` | — | Hollow block cursor (unfocused) |

## RenderStateColors

```go
type RenderStateColors struct {
    Background    ColorRGB
    Foreground    ColorRGB
    Cursor        ColorRGB
    CursorHasValue bool
    Palette       Palette
}
```

Retrieved in a single call via the sized-struct API. `Cursor` is only valid when `CursorHasValue` is `true`.

## Row and Cell Iteration

### RenderStateRow

Rows are iterated from the render state. Each row has:
- Dirty flag (whether this row changed since last update)
- Selection ranges (which cells are part of a selection)
- Cell iterator

### RenderStateCell

Cells are iterated from rows. Each cell provides:
- Resolved background/foreground colors
- Grapheme cluster (codepoints or UTF-8)
- Styling information
- Selection state
- Allocation-free grapheme extraction via `GraphemesInto`, `AppendGraphemes`

### RenderCellStyle

Cell-level style snapshot with resolved colors:
- `StyleInto` — Copy style into an existing `Style` value
- `ResolvedStyleInto` — Copy with resolved fg/bg colors

## Concurrency Rules

1. Hold **exclusive access** to the terminal while calling `Update`
2. After `Update` returns, the render state can be read **without** the terminal
3. Do **not** call `Update` concurrently with reads from the same render state
4. Iterators are valid only until the next `Update`
5. Copied values from getter methods can be retained freely

## Common Patterns

### Incremental Rendering

```go
rs, _ := libghostty.NewRenderState()
defer rs.Close()

for {
    err := rs.Update(term)
    if err != nil { break }

    dirty := rs.Dirty() // RenderStateDirtyFalse/Partial/Full
    if dirty == libghostty.RenderStateDirtyFalse {
        continue // Skip rendering
    }

    // Iterate rows and cells, only redraw dirty rows
    // ...
}
```

### Building a Custom Renderer

The render state is designed for building custom renderers (e.g., GPU-accelerated, web-based, or image-based). The pattern is:

1. Create a `RenderState`
2. On each frame, call `Update(term)` to sync
3. Check dirty state — skip if not dirty
4. Iterate rows/cells, extract graphemes and styles
5. Draw to your target surface
