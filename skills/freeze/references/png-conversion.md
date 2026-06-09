# PNG Conversion

Reference for how freeze converts SVG to PNG — the two fallback strategies and their tradeoffs.

## Overview

Freeze supports PNG output (via `--output file.png`) using a two-strategy fallback: first try `rsvg-convert` (fast, external), then fall back to `resvg-go` (slower, self-contained).

## Conversion Dispatch

```go
switch {
case strings.HasSuffix(config.Output, ".png"):
    svgConversionErr := libsvgConvert(doc, imageWidth, imageHeight, config.Output)
    if svgConversionErr == nil {
        printFilenameOutput(config.Output)
        break
    }
    svgConversionErr = resvgConvert(doc, imageWidth, imageHeight, config.Output)
    if svgConversionErr != nil {
        printErrorFatal("Unable to convert SVG to PNG", svgConversionErr)
    }
    printFilenameOutput(config.Output)
```

## Strategy 1: libsvgConvert (rsvg-convert)

```go
func libsvgConvert(doc *etree.Document, _, _ float64, output string) error
```

| Aspect | Detail |
|--------|--------|
| **Engine** | External `rsvg-convert` CLI |
| **Speed** | Fast |
| **Dependency** | Requires `rsvg-convert` on `$PATH` |
| **Font handling** | Delegated to `rsvg-convert` (uses system fonts) |
| **Dimensions** | Ignored (uses SVG intrinsic size) |
| **DPI** | Not exposed |

Process:
1. Check `exec.LookPath("rsvg-convert")` — return error if not found
2. Serialize etree document to SVG bytes
3. Pipe SVG bytes to `rsvg-convert -o <output>` via stdin
4. Return any error from the command

## Strategy 2: resvgConvert (resvg-go)

```go
func resvgConvert(doc *etree.Document, w, h float64, output string) error
```

| Aspect | Detail |
|--------|--------|
| **Engine** | `kanrichan/resvg-go` (Go bindings for resvg Rust library) |
| **Speed** | Slower (but self-contained) |
| **Dependency** | None (bundled) |
| **Font handling** | Explicit: loads embedded JetBrains Mono TTF fonts |
| **Dimensions** | Explicitly set via `w`, `h` |
| **DPI** | 192 (high-resolution) |

Process:
1. Serialize etree document to SVG bytes
2. Create `resvg.NewDefaultWorker(context.Background())`
3. Create font database: `worker.NewFontDBDefault()`
4. Load embedded fonts:
   - `font.JetBrainsMonoTTF` (regular)
   - `font.JetBrainsMonoNLTTF` (no ligatures)
5. Create pixmap: `worker.NewPixmap(uint32(w), uint32(h))`
6. Parse SVG tree with rendering options:
   - DPI: 192
   - Shape rendering: `GeometricPrecision`
   - Text rendering: `OptimizeLegibility`
   - Image rendering: `OptimizeQuality`
   - Default size: `w × h`
7. Convert text with font database: `tree.ConvertText(fontdb)`
8. Render to pixmap: `tree.Render(resvg.TransformIdentity(), pixmap)`
9. Encode PNG: `pixmap.EncodePNG()`
10. Write to file: `os.WriteFile(output, png, 0o600)`

All resources (worker, fontdb, pixmap, tree) are `defer`-closed.

## Scale Factor

When both width and height are auto and output is PNG:

```go
scale = 1
if autoHeight && autoWidth && strings.HasSuffix(config.Output, ".png") {
    scale = 4
}
```

The 4× scale produces high-resolution output suitable for retina/HiDPI displays. This scale is applied to:
- Margin and padding values
- Font size in `<text>` elements
- Image and terminal dimensions
- Shadow, border, and window control dimensions

## Comparison

| Aspect | `libsvgConvert` | `resvgConvert` |
|--------|----------------|----------------|
| Engine | External `rsvg-convert` CLI | Native `resvg-go` library |
| Speed | Faster | Slower |
| External dependency | Requires `rsvg-convert` on PATH | None |
| Font handling | Delegated to system | Explicit (JetBrains Mono) |
| DPI control | Not exposed | 192 DPI |
| Rendering options | Not exposed | Full control |
| Dimensions | Uses SVG intrinsic size | Explicitly set |
| Error handling | Returns error (triggers fallback) | Fatal on worker/fontdb errors |

## Gotchas

- **WebP is not supported**: Despite `--output` help text mentioning `.webp`, freeze v0.2.2 produces SVG when given a `.webp` extension. The PNG conversion path only triggers for `.png` suffix.
- **Custom fonts may not render in resvgConvert**: Only JetBrains Mono is loaded into the font database. If `--font.family` is set to a non-embedded font, text may render with a fallback font or not at all.
- **`rsvg-convert` uses system fonts**: It resolves fonts from the system, so custom fonts specified via `--font.family` may work if installed system-wide.
