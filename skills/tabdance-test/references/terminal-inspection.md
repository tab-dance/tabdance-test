# Terminal Inspection — Reading Terminal State

Reference for inspecting headless terminal state in tabdance-test: screen content, cursor, cells, and styles.

## Overview

tabdance-test provides several MCP tools for reading terminal state, all backed by go-libghostty's Formatter, GridRef, and data getter APIs.

## get_view — Full Screen Content

Uses `Formatter` to extract the entire visible screen:

```go
func (m *ManagedTerminal) GetView(format string, trim, unwrap bool) (ViewResult, error) {
    m.mu.Lock()
    defer m.mu.Unlock()

    var fmtFormat libghostty.FormatterFormat
    switch format {
    case "plain":
        fmtFormat = libghostty.FormatterFormatPlain
    case "vt":
        fmtFormat = libghostty.FormatterFormatVT
    case "html":
        fmtFormat = libghostty.FormatterFormatHTML
    }

    f, _ := libghostty.NewFormatter(m.Terminal,
        libghostty.WithFormatterFormat(fmtFormat),
        libghostty.WithFormatterTrim(trim),
        libghostty.WithFormatterUnwrap(unwrap),
    )
    defer f.Close()

    content, _ := f.FormatString()

    return ViewResult{
        Content:      content,
        CursorX:      m.Terminal.CursorX(),
        CursorY:      m.Terminal.CursorY(),
        CursorVisible: m.Terminal.CursorVisible(),
        ActiveScreen: m.Terminal.ActiveScreen(),
    }, nil
}
```

## get_cell — Individual Cell Inspection

Uses `GridRef` to inspect a specific cell:

```go
func (m *ManagedTerminal) GetCell(x uint16, y uint32) (CellResult, error) {
    m.mu.Lock()
    defer m.mu.Unlock()

    point := libghostty.Point{
        Tag: libghostty.PointTagViewport,
        X:   x,
        Y:   y,
    }

    ref, _ := m.Terminal.GridRef(point)
    cell := ref.Cell()
    style := ref.Style()

    result := CellResult{
        Codepoint: cell.Codepoint(),
        Wide:      cell.Wide(),
        Bold:      style.Bold(),
        Italic:    style.Italic(),
        FgColor:   style.FgColor().String(),
        BgColor:   style.BgColor().String(),
    }

    if graphemes := ref.Graphemes(); len(graphemes) > 0 {
        result.Text = string(graphemes)
    }

    if uri := ref.HyperlinkURI(); uri != "" {
        result.Hyperlink = uri
    }

    return result, nil
}
```

## get_cursor — Cursor State

```go
func (m *ManagedTerminal) GetCursor() CursorResult {
    m.mu.Lock()
    defer m.mu.Unlock()

    style := m.Terminal.CursorStyle()

    var cursorStyle string
    switch {
    case style.Bold():
        cursorStyle = "block"
    default:
        cursorStyle = "bar"
    }

    return CursorResult{
        X:           m.Terminal.CursorX(),
        Y:           m.Terminal.CursorY(),
        Visible:     m.Terminal.CursorVisible(),
        PendingWrap: m.Terminal.CursorPendingWrap(),
        Style:       cursorStyle,
    }
}
```

## get_size — Terminal Dimensions

```go
func (m *ManagedTerminal) GetSize() SizeResult {
    return SizeResult{
        Cols:     m.Terminal.Cols(),
        Rows:     m.Terminal.Rows(),
        WidthPx:  m.Terminal.WidthPx(),
        HeightPx: m.Terminal.HeightPx(),
    }
}
```

## Advanced Inspection

### Scrollback Access

```go
// Access scrollback history
point := libghostty.Point{
    Tag: libghostty.PointTagHistory,
    X:   0,
    Y:   0,  // First row of scrollback
}
ref, _ := m.Terminal.GridRef(point)
```

### Selection Content

```go
if sel := m.Terminal.Selection(); sel != nil {
    text, _ := m.Terminal.SelectionFormatString(
        libghostty.WithSelectionFormat(libghostty.FormatterFormatPlain),
    )
}
```

### Semantic Content (Shell Integration)

```go
// Check if a cell is prompt, input, or output
point := libghostty.Point{Tag: libghostty.PointTagViewport, X: 0, Y: 5}
ref, _ := m.Terminal.GridRef(point)
cell := ref.Cell()
semantic := cell.Semantic()
// CellSemanticContentPrompt, CellSemanticContentInput, CellSemanticContentOutput
```

### Mode Checking

```go
// Check if bracketed paste is active
bp, _ := m.Terminal.ModeGet(libghostty.ModeBracketedPaste)

// Check if mouse tracking is active
mt := m.Terminal.MouseTracking()

// Check Kitty keyboard flags
kkf := m.Terminal.KittyKeyboardFlags()
```

## Render State for Visual Testing

For pixel-level visual testing, use `RenderState`:

```go
rs, _ := libghostty.NewRenderState()
defer rs.Close()

rs.Update(m.Terminal)

dirty := rs.Dirty()
if dirty == libghostty.RenderStateDirtyFalse {
    // No changes since last update
}

// Access colors
colors := rs.Colors()
// colors.Background, colors.Foreground, colors.Palette
```
