# Terminal Inspection — Reading Terminal State

Reference for inspecting headless terminal state in tabdance-test: screen content, cursor, cells, and styles.

## Overview

tabdance-test provides several MCP tools for reading terminal state, all backed by go-libghostty's Formatter, GridRef, and data getter APIs. All Terminal data getters return `(value, error)` tuples.

## get_view — Full Screen Content

Uses `Formatter` to extract the entire visible screen, plus cursor state and active screen:

```go
type ViewResult struct {
    Content       string `json:"content"`
    CursorX       uint16 `json:"cursor_x"`
    CursorY       uint16 `json:"cursor_y"`
    CursorVisible bool   `json:"cursor_visible"`
    CursorStyle   string `json:"cursor_style"`
    ActiveScreen  string `json:"active_screen"`
}

func GetView(term *libghostty.Terminal, format string, trim, unwrap bool) (*ViewResult, error) {
    f, _ := libghostty.NewFormatter(term,
        libghostty.WithFormatterFormat(fmtFormat),
        libghostty.WithFormatterTrim(trim),
        libghostty.WithFormatterUnwrap(unwrap),
    )
    defer f.Close()

    content, _ := f.FormatString()

    cursorX, _ := term.CursorX()
    cursorY, _ := term.CursorY()
    cursorVisible, _ := term.CursorVisible()
    screen, _ := term.ActiveScreen()

    styleStr := "bar"
    if cs, err := term.CursorStyle(); err == nil && cs != nil {
        if cs.Bold() {
            styleStr = "block"
        }
    }

    return &ViewResult{
        Content:       content,
        CursorX:       cursorX,
        CursorY:       cursorY,
        CursorVisible: cursorVisible,
        CursorStyle:   styleStr,
        ActiveScreen:  screenStr,
    }, nil
}
```

## get_cell — Individual Cell Inspection

Uses `GridRef` to inspect a specific cell:

```go
type CellResult struct {
    Codepoint uint32 `json:"codepoint"`
    Text      string `json:"text"`
    Wide      bool   `json:"wide"`
    Bold      bool   `json:"bold"`
    Italic    bool   `json:"italic"`
    FgColor   string `json:"fg"`
    BgColor   string `json:"bg"`
    Hyperlink string `json:"hyperlink,omitempty"`
}

func GetCell(term *libghostty.Terminal, x uint16, y uint32) (*CellResult, error) {
    point := libghostty.Point{
        Tag: libghostty.PointTagViewport,
        X:   x,
        Y:   y,
    }

    ref, _ := term.GridRef(point)

    // Cell data
    cell, _ := ref.Cell()
    cp, _ := cell.Codepoint()
    wide, _ := cell.Wide()
    // wide == libghostty.CellWideWide means wide character

    // Graphemes (returns []uint32 codepoints, not string)
    graphemes, _ := ref.Graphemes()
    runes := make([]rune, len(graphemes))
    for i, cp := range graphemes {
        runes[i] = rune(cp)
    }
    text := string(runes)

    // Style
    style, _ := ref.Style()
    bold := style.Bold()
    italic := style.Italic()
    fgColor := styleColorString(style.FgColor())
    bgColor := styleColorString(style.BgColor())

    // Hyperlink
    uri, _ := ref.HyperlinkURI()

    return &CellResult{...}, nil
}
```

### StyleColor Formatting

```go
func styleColorString(sc libghostty.StyleColor) string {
    switch sc.Tag {
    case libghostty.StyleColorNone:
        return "default"
    case libghostty.StyleColorPalette:
        return fmt.Sprintf("color%d", sc.Palette)
    case libghostty.StyleColorRGB:
        return fmt.Sprintf("#%02x%02x%02x", sc.RGB.R, sc.RGB.G, sc.RGB.B)
    default:
        return "default"
    }
}
```

## get_cursor — Cursor State

```go
type CursorResult struct {
    X           uint16 `json:"x"`
    Y           uint16 `json:"y"`
    Visible     bool   `json:"visible"`
    PendingWrap bool   `json:"pending_wrap"`
    Style       string `json:"style"`
}

func GetCursor(term *libghostty.Terminal) *CursorResult {
    x, _ := term.CursorX()
    y, _ := term.CursorY()
    visible, _ := term.CursorVisible()
    pendingWrap, _ := term.CursorPendingWrap()

    styleStr := "bar"
    if cs, err := term.CursorStyle(); err == nil && cs != nil {
        if cs.Bold() {
            styleStr = "block"
        }
    }

    return &CursorResult{X: x, Y: y, Visible: visible, PendingWrap: pendingWrap, Style: styleStr}
}
```

## get_size — Terminal Dimensions

```go
type SizeResult struct {
    Cols     uint16 `json:"cols"`
    Rows     uint16 `json:"rows"`
    WidthPx  uint32 `json:"width_px"`
    HeightPx uint32 `json:"height_px"`
}

func GetSize(term *libghostty.Terminal) *SizeResult {
    cols, _ := term.Cols()
    rows, _ := term.Rows()
    widthPx, _ := term.WidthPx()
    heightPx, _ := term.HeightPx()
    return &SizeResult{Cols: cols, Rows: rows, WidthPx: widthPx, HeightPx: heightPx}
}
```

## Advanced Inspection

### Scrollback Access

```go
point := libghostty.Point{
    Tag: libghostty.PointTagHistory,
    X:   0,
    Y:   0,  // First row of scrollback
}
ref, _ := term.GridRef(point)
```

### Mode Checking

```go
// Check if bracketed paste is active
bp, _ := term.ModeGet(libghostty.ModeBracketedPaste)

// Check Kitty keyboard flags
kkf, _ := term.KittyKeyboardFlags()
```

### CellWide Constants

- `CellWideNarrow` — normal character
- `CellWideWide` — wide character (CJK, emoji)
- `CellWideSpacerHead` — first cell of a wide character
- `CellWideSpacerTail` — second cell of a wide character
