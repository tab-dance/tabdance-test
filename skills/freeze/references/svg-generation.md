# SVG Generation

Reference for how freeze constructs SVG output using Chroma and etree.

## Overview

Freeze generates SVG by: (1) tokenizing input with Chroma, (2) formatting tokens to SVG with Chroma's SVG formatter, (3) parsing the result as XML with `beevik/etree`, and (4) post-processing the XML document to apply layout, decorations, and ANSI styling.

## Pipeline

```
Input string
  → Chroma lexer.Tokenise() → chroma.Iterator
    → Chroma SVG formatter.Format() → bytes.Buffer (SVG XML)
      → etree.NewDocument().ReadFrom(buf) → *etree.Document
        → Post-processing: dimensions, padding, margin, decorations
          → Final SVG document
```

## Chroma Tokenization

For non-ANSI input:

```go
lexer := lexers.Get(config.Input)       // by filename
// or
lexer := lexers.Get(config.Language)    // by --language flag
// or
lexer := lexers.Analyse(input)          // auto-detect from content

it, err := chroma.Coalesce(lexer).Tokenise(nil, input)
```

For ANSI input, a single `Text` token is used (see [ansi-rendering.md](ansi-rendering.md)).

## Chroma SVG Formatter

```go
options, err := fontOptions(&config)
f := formatter.New(options...)
buf := &bytes.Buffer{}
err = f.Format(buf, style, it)
```

The formatter produces an SVG document with:
- An outer `<svg>` element with width/height
- A `<rect>` for the terminal background
- A `<g>` group containing `<text>` elements per line
- Each `<text>` contains `<tspan>` elements for styled segments

## Font Options

The `fontOptions` function determines how fonts are embedded in the SVG:

| Condition | Font Handling |
|-----------|---------------|
| `--font.file` is set | Read TTF/WOFF/WOFF2 file, base64-encode, embed as `@font-face` in SVG `<style>` |
| `--font.family` is not "JetBrains Mono" | Set `font-family` only (no embedding — user must have the font) |
| Default (JetBrains Mono) | Embed base64-encoded WOFF2 from `font.JetBrainsMono` or `font.JetBrainsMonoNL` (if `--font.ligatures=false`) |

```go
func fontOptions(config *Config) ([]svg.Option, error) {
    if config.Font.File != "" {
        // Read and embed custom font
        return []svg.Option{
            svg.EmbedFont(config.Font.Family, base64Enc, format),
            svg.FontFamily(config.Font.Family),
        }, nil
    }
    if config.Font.Family != "JetBrains Mono" {
        return []svg.Option{svg.FontFamily(config.Font.Family)}, nil
    }
    // Default: embed JetBrains Mono WOFF2
    return []svg.Option{
        svg.EmbedFont("JetBrains Mono", font.JetBrainsMono, svg.WOFF2),
        svg.FontFamily("JetBrains Mono"),
    }, nil
}
```

Supported font file formats: `.ttf` (TRUETYPE), `.woff2` (WOFF2), `.woff` (WOFF).

## XML Post-Processing

After Chroma produces the initial SVG, freeze modifies the etree document:

### 1. Dimension Calculation

```go
w, h := svg.GetDimensions(image)
imageWidth := float64(w) * scale
imageHeight := float64(h) * scale
// Adjust for font size and line height
imageHeight *= (config.Font.Size / defaultFontSize)
imageHeight *= (config.LineHeight / defaultLineHeight)
```

### 2. Auto-Width Calculation

When width is auto-sized:

```go
longestLine := lipgloss.Width(strings.ReplaceAll(strippedInput, "\t", strings.Repeat(" ", tabWidth)))
terminalWidth = float64(longestLine+1) * (config.Font.Size / fontHeightToWidthRatio)
terminalWidth *= scale
terminalWidth += hPadding
imageWidth = terminalWidth + hMargin
```

### 3. Text Repositioning

Each `<text>` line is repositioned with padding and margin offsets:

```go
x := float64(config.Padding[left] + config.Margin[left])
y := (float64(i+1))*(config.Font.Size*config.LineHeight) + float64(config.Padding[top]) + config.Margin[top]
svg.Move(line, x, y)
```

Lines beyond the visible area are removed.

### 4. Decorations Applied

| Decoration | How |
|------------|-----|
| Window controls | `svg.NewWindowControls()` → added as child of `<svg>` |
| Border radius | `svg.AddCornerRadius(terminal, radius)` |
| Shadow | `svg.AddShadow()` → filter applied to `<rect>` |
| Border outline | `svg.AddOutline(terminal, width, color)` |
| Line numbers | `<tspan>` prepended to each `<text>` |
| Clip path | `svg.AddClipPath()` when fixed width/height |

### 5. ANSI Processing

If input is ANSI, the dispatcher runs after text repositioning (see [ansi-rendering.md](ansi-rendering.md)).

## Constants

```go
const (
    defaultFontSize   = 14.0
    defaultLineHeight = 1.2
)
const fontHeightToWidthRatio = 1.68
```
