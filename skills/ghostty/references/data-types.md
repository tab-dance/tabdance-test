# Data Types — Core Types and Constants

In-depth reference for the core data types in go-libghostty: Cell, Row, Style, Color, Mode, DeviceAttributes, SizeReport, BuildInfo, Result, and Error.

## Style — Cell Styling

```go
type Style struct { /* internal */ }
```

A copied snapshot of a cell's SGR styling. Can be retained freely.

### Methods

| Method | Return | Description |
|--------|--------|-------------|
| `IsDefault()` | `bool` | Whether style is all defaults |
| `FgColor()` | `StyleColor` | Foreground color |
| `BgColor()` | `StyleColor` | Background color |
| `UnderlineColor()` | `StyleColor` | Underline color |
| `Bold()` | `bool` | Bold attribute |
| `Italic()` | `bool` | Italic attribute |
| `Faint()` | `bool` | Faint/dim attribute |
| `Blink()` | `bool` | Blink attribute |
| `Inverse()` | `bool` | Inverse/reverse video |
| `Invisible()` | `bool` | Invisible/hidden attribute |
| `Strikethrough()` | `bool` | Strikethrough attribute |
| `Overline()` | `bool` | Overline attribute |
| `Underline()` | `UnderlineConstant` | Underline style |

## StyleColor — Tagged Color Union

```go
type StyleColor int
```

| Constant | Description |
|----------|-------------|
| `StyleColorNone` | No color (default) |
| `StyleColorPalette` | Palette index (0–255) |
| `StyleColorRGB` | True-color RGB |

## ColorRGB — True Color

```go
type ColorRGB struct { R, G, B uint8 }
```

24-bit RGB color value. Regular Go value, can be retained freely.

## Palette — 256-Color Palette

```go
type Palette [256]ColorRGB
```

Full 256-color palette. Regular Go value.

## Mode — Terminal Modes

```go
type Mode uint32
```

ANSI and DEC terminal modes. Key constants:

| Constant | Description |
|----------|-------------|
| `ModeCursorKeyApplication` | DEC mode 1 |
| `ModeBracketedPaste` | DEC mode 2004 |
| `ModeFocusEvent` | DEC mode 1004 |
| `ModeMouseX10` | DEC mode 9 |
| `ModeMouseNormal` | DEC mode 1000 |
| `ModeMouseButton` | DEC mode 1002 |
| `ModeMouseAnyEvent` | DEC mode 1003 |
| `ModeMouseSGR` | DEC mode 1006 |
| `ModeMouseURxvt` | DEC mode 1015 |
| `ModeMouseSGRPixels` | DEC mode 1016 |
| `ModeAlternateScreen` | DEC mode 1049 |
| `ModeAutoWrap` | DEC mode 7 |
| `ModeInsert` | DEC mode 4 |
| `ModeOrigin` | DEC mode 6 |
| `ModeLineFeedNewLine` | DEC mode 20 |
| `ModeVisibleCursor` | DEC mode 25 |
| `ModeScrollRegion` | DEC mode 3 |

Full list available via `ghostty +list-actions` or the C header `ghostty/vt/modes.h`.

## ModeReportState

```go
type ModeReportState int
```

State of a mode for DECRQM/DECRPM reporting.

## DeviceAttributes

```go
type DeviceAttributes struct { /* internal */ }
```

Device attributes for DA1/DA2/DA3 queries. Used with `WithDeviceAttributes` effect callback.

## SizeReportSize

```go
type SizeReportSize struct {
    WidthPx  uint32
    HeightPx uint32
}
```

Used with `WithSizeReport` effect callback for XTWINOPS queries.

## ColorScheme

```go
type ColorScheme int
```

| Constant | Description |
|----------|-------------|
| `ColorSchemeDark` | Dark color scheme |
| `ColorSchemeLight` | Light color scheme |
| `ColorSchemeNoPreference` | No preference |

## BuildInfo

```go
type BuildInfo struct {
    SIMD          bool
    KittyGraphics bool
    TmuxControlMode bool
    Optimize      string
    VersionString string
    VersionMajor  uint32
    VersionMinor  uint32
    VersionPatch  uint32
    VersionPre    string
    VersionBuild  string
}
```

```go
func GetBuildInfo() (BuildInfo, error)
```

Returns compile-time build configuration of the linked libghostty library.

## Result — Error Codes

```go
type Result int

const (
    ResultSuccess     Result = C.GHOSTTY_SUCCESS
    ResultOutOfMemory Result = C.GHOSTTY_OUT_OF_MEMORY
    ResultInvalidValue Result = C.GHOSTTY_INVALID_VALUE
    ResultOutOfSpace  Result = C.GHOSTTY_OUT_OF_SPACE
    ResultNoValue     Result = C.GHOSTTY_NO_VALUE
)
```

## Error

```go
type Error struct { Result Result }

func (e *Error) Error() string
```

All C API calls go through `resultError()` which converts non-success results to Go errors.

## Underline Constants

| Constant | Description |
|----------|-------------|
| `UnderlineNone` | No underline |
| `UnderlineSingle` | Single underline |
| `UnderlineDouble` | Double underline |
| `UnderlineCurly` | Curly underline |
| `UnderlineDotted` | Dotted underline |
| `UnderlineDashed` | Dashed underline |

## Text Serialization

`Key`, `Mods`, `MouseButton`, `FocusEvent` all implement:
- `String()` → canonical snake_case name
- `encoding.TextMarshaler` / `encoding.TextUnmarshaler`
- Top-level `ParseKey`, `ParseMods`, `ParseMouseButton`, `ParseFocusEvent` functions

This enables JSON, TOML, YAML, XML serialization and `flag.TextVar` usage.
