# Grid Reference — Cell and Row Inspection

In-depth reference for `GridRef`, `TrackedGridRef`, and `Point` in go-libghostty: inspecting individual cells and rows in the terminal grid.

## Overview

Grid references provide access to individual cells and rows in the terminal grid. They support multiple coordinate systems and can track cells across screen updates.

## Point — Grid Coordinates

```go
type Point struct {
    Tag PointTag
    X   uint16
    Y   uint32
}
```

### PointTag — Coordinate Systems

```go
type PointTag int

const (
    PointTagActive  PointTag = C.GHOSTTY_POINT_TAG_ACTIVE   // Active area (cursor-relative)
    PointTagViewport PointTag = C.GHOSTTY_POINT_TAG_VIEWPORT // Visible viewport
    PointTagScreen  PointTag = C.GHOSTTY_POINT_TAG_SCREEN    // Full screen (primary + scrollback)
    PointTagHistory PointTag = C.GHOSTTY_POINT_TAG_HISTORY   // Scrollback history only
)
```

| Tag | Description | Performance |
|-----|-------------|-------------|
| `PointTagActive` | Active area (bottom of screen) | Fast |
| `PointTagViewport` | Visible viewport | Fast |
| `PointTagScreen` | Full screen including scrollback | May be expensive for large scrollback |
| `PointTagHistory` | Scrollback history only | May be expensive for large scrollback |

### PointCoordinate

```go
type PointCoordinate struct {
    X uint16
    Y uint32
}
```

## GridRef — Borrowed Grid Reference

```go
type GridRef struct { ref C.GhosttyGridRef }
```

A borrowed view into the terminal grid. **Only valid until the next mutating terminal call** (VTWrite, Reset, Resize, etc.).

### Obtaining a GridRef

```go
func (t *Terminal) GridRef(point Point) (*GridRef, error)
```

### GridRef Methods

| Method | Return | Description |
|--------|--------|-------------|
| `Cell()` | `Cell` | Copied cell snapshot |
| `Row()` | `Row` | Copied row snapshot |
| `Graphemes()` | `[]rune` | Full grapheme cluster codepoints |
| `HyperlinkURI()` | `string` | Hyperlink URI (empty if none) |
| `Style()` | `Style` | Copied cell style |

### Converting Back to Point

```go
func (t *Terminal) PointFromGridRef(ref *GridRef, tag PointTag) (Point, error)
```

Converts a grid reference back to a point in the requested coordinate system. Returns `ResultNoValue` if the reference cannot be represented in that coordinate system.

## TrackedGridRef — Owned Grid Reference

```go
type TrackedGridRef struct { ptr C.GhosttyTrackedGridRef }
```

An owned reference that **follows a cell through screen updates** (scrolling, pruning, resize/reflow). Must be closed when finished.

### Obtaining a TrackedGridRef

```go
func (t *Terminal) TrackGridRef(point Point) (*TrackedGridRef, error)
```

### Lifecycle

```go
func (ref *TrackedGridRef) Close()
```

### Behavior

- Attached to the terminal screen/page-list that is active when created
- If the terminal is closed first, the tracked handle remains valid only for tracked-grid-ref APIs: it reports no value and can still be closed
- Useful for tracking cursor position or specific cells across updates

## Cell — Copied Cell Snapshot

```go
type Cell struct { /* internal */ }
```

A copied snapshot of a single terminal cell. Can be retained freely.

### Cell Methods

| Method | Return | Description |
|--------|--------|-------------|
| `Codepoint()` | `uint32` | Cell codepoint (0 if empty) |
| `ContentTag()` | `CellContentTag` | Content type |
| `Wide()` | `CellWide` | Width property |
| `HasText()` | `bool` | Whether cell has text |
| `HasStyling()` | `bool` | Whether cell has non-default styling |
| `StyleID()` | `uint16` | Style ID for lookups |
| `HasHyperlink()` | `bool` | Whether cell has a hyperlink |
| `Protected()` | `bool` | Whether cell is protected |
| `Semantic()` | `CellSemanticContent` | Semantic type (from OSC 133) |
| `ColorPalette()` | `ColorPaletteIndex` | Palette index for bg color |
| `ColorRGB()` | `ColorRGB` | RGB value for bg color |
| `GetMulti(keys, values)` | `error` | Batch read |

### CellContentTag

| Tag | Description |
|-----|-------------|
| `CellContentTagCodepoint` | Simple codepoint |
| `CellContentTagGrapheme` | Grapheme cluster |
| `CellContentTagBgColorPalette` | Background is palette color |
| `CellContentTagBgColorRGB` | Background is RGB color |

### CellWide

| Value | Description |
|-------|-------------|
| `CellWideNarrow` | Single-width cell |
| `CellWideWide` | Wide character (occupies 2 cells) |
| `CellWideSpacerTail` | Second cell of wide character |
| `CellWideSpacerHead` | First cell of wide character |

## Row — Copied Row Snapshot

```go
type Row struct { /* internal */ }
```

A copied snapshot of a single terminal row. Can be retained freely.

### Row Methods

| Method | Return | Description |
|--------|--------|-------------|
| `Wrap()` | `bool` | Whether row is soft-wrapped |
| `WrapContinuation()` | `bool` | Whether row is a continuation |
| `Grapheme()` | `bool` | Whether any cells have grapheme clusters |
| `Styled()` | `bool` | Whether any cells have styling |
| `Hyperlink()` | `bool` | Whether any cells have hyperlinks |
| `Semantic()` | `RowSemanticPrompt` | Prompt state |
| `KittyVirtualPlaceholder()` | `bool` | Kitty virtual placeholder |
| `Dirty()` | `bool` | Whether row needs redraw |
| `GetMulti(keys, values)` | `error` | Batch read |

## Common Patterns

### Inspect Cell at Cursor Position

```go
point := libghostty.Point{
    Tag: libghostty.PointTagActive,
    X:   term.CursorX(),
    Y:   uint32(term.CursorY()),
}
ref, _ := term.GridRef(point)
cell := ref.Cell()
style := ref.Style()
graphemes := ref.Graphemes()
```

### Iterate Visible Viewport

```go
for y := uint32(0); y < uint32(term.Rows()); y++ {
    for x := uint16(0); x < term.Cols(); x++ {
        point := libghostty.Point{
            Tag: libghostty.PointTagViewport,
            X:   x,
            Y:   y,
        }
        ref, err := term.GridRef(point)
        if err != nil { continue }
        cell := ref.Cell()
        if cell.HasText() {
            // Process cell
        }
    }
}
```

### Track a Cell Across Updates

```go
point := libghostty.Point{
    Tag: libghostty.PointTagActive,
    X:   term.CursorX(),
    Y:   uint32(term.CursorY()),
}
tracked, _ := term.TrackGridRef(point)
defer tracked.Close()

// ... feed more VT data ...
term.VTWrite(moreData)

// The tracked reference still points to the same logical cell
// even if it moved due to scrolling or reflow
```
