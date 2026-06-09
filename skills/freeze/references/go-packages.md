# Go Packages

Reference for the importable Go packages in the freeze repository (svg, input, font).

## Overview

Freeze's `main` package is not importable, but three sub-packages provide utility functions that can be used independently.

## `github.com/charmbracelet/freeze/svg` — SVG Element Manipulation

All functions operate on `*etree.Element` (from [beevik/etree](https://github.com/beevik/etree), an XML library).

| Function | Signature | Description |
|----------|-----------|-------------|
| `AddShadow` | `(element *etree.Element, id string, x, y, blur float64)` | Add SVG shadow filter definition to `<defs>` |
| `AddClipPath` | `(element *etree.Element, id string, x, y, w, h float64)` | Add SVG clip path definition to `<defs>` |
| `AddCornerRadius` | `(e *etree.Element, radius float64)` | Add `rx`/`ry` corner radius attributes |
| `Move` | `(e *etree.Element, x, y float64)` | Set element position via `x`/`y` attributes |
| `AddOutline` | `(e *etree.Element, width float64, color string)` | Add stroke outline to element |
| `NewWindowControls` | `(r float64, x, y float64) *etree.Element` | Create macOS-style window dots (red/yellow/green) |
| `SetDimensions` | `(element *etree.Element, width, height float64)` | Set `width`/`height` attributes |
| `GetDimensions` | `(element *etree.Element) (int, int)` | Read `width`/`height` from element |

### Constants

```go
const (
    red    = "#FF5A54"
    yellow = "#E6BF29"
    green  = "#52C12B"
)
```

Used by `NewWindowControls` for the window dot colors.

## `github.com/charmbracelet/freeze/input` — Input Reading Helpers

| Function | Signature | Description |
|----------|-----------|-------------|
| `ReadFile` | `(file string) (string, error)` | Read file contents as string |
| `ReadInput` | `(in io.Reader) (string, error)` | Read all data from an `io.Reader` |
| `IsPipe` | `(in *os.File) bool` | Check if stdin is a pipe (not a TTY) |

`IsPipe` checks for the absence of `os.ModeCharDevice`:

```go
func IsPipe(in *os.File) bool {
    stat, err := in.Stat()
    if err != nil {
        return false
    }
    return (stat.Mode() & os.ModeCharDevice) == 0
}
```

## `github.com/charmbracelet/freeze/font` — Embedded Font Data

| Variable | Type | Description |
|----------|------|-------------|
| `JetBrainsMonoTTF` | `[]byte` | Raw TTF font bytes (embedded via `go:embed`) |
| `JetBrainsMonoNLTTF` | `[]byte` | No-ligatures variant TTF bytes |
| `JetBrainsMono` | `string` | Base64-encoded TTF (regular) |
| `JetBrainsMonoNL` | `string` | Base64-encoded TTF (no ligatures) |

The TTF bytes are embedded at compile time from `JetBrainsMono-Regular.ttf` and `JetBrainsMonoNL-Regular.ttf`. The base64 string variants are pre-encoded for direct use in SVG `<style>` `@font-face` declarations.

## Limitations

These packages are low-level building blocks — they don't provide a "generate an image from code" workflow on their own. The main image generation logic (tokenization, layout, ANSI handling, PNG conversion) lives entirely in `package main` and is not importable.

For programmatic usage, shell out to the `freeze` CLI (see [programmatic.md](programmatic.md)).