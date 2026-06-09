# ANSI Rendering

Reference for how freeze renders ANSI escape sequences in SVG output.

## Overview

When freeze detects ANSI input (either via `--language ansi` or by detecting escape sequences in the input), it uses a custom `dispatcher` to parse ANSI sequences and inject styled SVG elements directly into the XML document. This bypasses Chroma's tokenization entirely.

## Detection

```go
strippedInput := ansi.Strip(input)
isAnsi := strings.ToLower(config.Language) == "ansi" || strippedInput != input
```

Input is treated as ANSI if:
- `--language ansi` is explicitly set, OR
- The input contains ANSI escape sequences (stripped content differs from original)

## Tokenization Bypass

For ANSI input, instead of Chroma tokenization:

```go
if isAnsi {
    it = chroma.Literator(chroma.Token{Type: chroma.Text, Value: strippedInput})
}
```

A single `Text` token containing the stripped (plain) text is created. This is used only for initial SVG sizing — the actual styled content comes from the dispatcher.

## Dispatcher

The `dispatcher` struct processes ANSI sequences and modifies SVG elements in-place:

```go
type dispatcher struct {
    scale   float64
    svg     *etree.Element
    bg      *etree.Element
    config  *Config
    lines   []*etree.Element
    row     int
    col     int
    bgWidth int
}
```

### Handler Interface

The dispatcher implements `ansi.Handler` with three methods:

| Method | Purpose |
|--------|---------|
| `Print(r rune)` | Insert a character into the current line's last `<tspan>` |
| `Execute(code byte)` | Handle control characters (tab, newline) |
| `CsiDispatch(cmd ansi.Cmd, params ansi.Params)` | Handle CSI sequences (only `m` = SGR) |

### Processing Flow

```go
parser := ansi.NewParser()
parser.SetHandler(ansi.Handler{
    Print:     d.Print,
    HandleCsi: d.CsiDispatch,
    Execute:   d.Execute,
})
for _, line := range strings.Split(input, "\n") {
    parser.Parse([]byte(line))
    d.Execute(ansi.LF)
}
```

Each line is parsed separately, with a simulated newline (`LF`) after each.

## SGR (Select Graphic Rendition) Support

The `CsiDispatch` method handles CSI `m` sequences:

| SGR Code | SVG Effect |
|----------|-----------|
| `0` | Reset — creates new empty `<tspan>` (clears styles) |
| `1` | Bold — adds `<tspan>` (note: `font-weight: bold` is commented out) |
| `3` | Italic — `font-style: italic` |
| `4` | Underline — `text-decoration: underline` |
| `9` | Strikethrough — `text-decoration: line-through` |
| `30`–`37` | Standard foreground colors — `fill` attribute |
| `38;5;n` | 256-color foreground — `fill` attribute |
| `38;2;r;g;b` | True-color foreground — `fill="#RRGGBB"` |
| `48;5;n` | 256-color background — `<rect>` element |
| `48;2;r;g;b` | True-color background — `<rect>` element |
| `90`–`97` | Bright foreground colors — `fill` attribute |
| `100`–`107` | Bright background colors — `<rect>` element |

Non-`m` CSI sequences are silently ignored.

## Foreground Color Palette

Standard and bright foreground colors:

```go
var ansiPalette = map[int]string{
    30: "#282a2e", 31: "#D74E6F", 32: "#31BB71", 33: "#D3E561",
    34: "#8056FF", 35: "#ED61D7", 36: "#04D7D7", 37: "#C5C8C6",
    90: "#4B4B4B", 91: "#FE5F86", 92: "#00D787", 93: "#EBFF71",
    94: "#8F69FF", 95: "#FF7AEA", 96: "#00FEFE", 97: "#FFFFFF",
}
```

## Background Rendering

Background colors are rendered as `<rect>` elements inserted before the text group:

1. `beginBackground(fill)` — creates a `<rect>` with the fill color, positioned at the current (col, row)
2. Characters increment `bgWidth` as they're printed
3. `endBackground()` — calculates width from `bgWidth`, inserts the `<rect>` at position 0 in the SVG (behind text)

The background rect height is `font.Size * lineHeight + 1` px.

## Wide Character Handling

```go
if runewidth.RuneWidth(r) > 1 {
    newChild := lastChild.Copy()
    newChild.SetText(string(r))
    newChild.CreateAttr("dx", fmt.Sprintf("%.2fpx", (p.config.Font.Size/5)*p.scale))
    p.lines[p.row].AddChild(newChild)
}
```

Wide characters (CJK, emoji) are placed in a separate `<tspan>` with a `dx` offset to prevent overlap.

## Tab Handling

```go
func (p *dispatcher) Execute(code byte) {
    if code == '\t' {
        for p.col%16 != 0 {
            p.Print(' ')
        }
    }
    if code == '\n' {
        p.endBackground()
        p.row++
        p.col = 0
    }
}
```

Tabs expand to spaces, aligned to 16-column boundaries (tab width = 6 for ANSI mode).

## Font Height-to-Width Ratio

```go
const fontHeightToWidthRatio = 1.68
```

Used for calculating character positions and background rect widths.
