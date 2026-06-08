# Formatter — Terminal Content Extraction

In-depth reference for `Formatter` in go-libghostty: extracting terminal content as plain text, VT sequences, or HTML.

## Overview

`Formatter` produces a snapshot of the terminal's active screen content in one of three formats. It borrows a reference to the terminal, so the terminal must outlive the formatter and all calls must be serialized with terminal access.

```go
f, _ := libghostty.NewFormatter(term,
    libghostty.WithFormatterFormat(libghostty.FormatterFormatPlain),
    libghostty.WithFormatterTrim(true),
)
defer f.Close()

content, _ := f.FormatString()
// content = "Hello, world!" (stripped of escape sequences)
```

## Formatter

```go
type Formatter struct { ptr C.GhosttyFormatter }
```

### Lifecycle

```go
func NewFormatter(t *Terminal, opts ...FormatterOption) (*Formatter, error)
func (f *Formatter) Close()
```

### Format Methods

```go
func (f *Formatter) Format() ([]byte, error)
func (f *Formatter) FormatString() (string, error)
func (f *Formatter) WriteTo(w io.Writer) (int64, error)  // implements io.WriterTo
```

Each call reflects the terminal's current state at the time of the call. Serialize `Format` with all other access to the underlying terminal.

## FormatterFormat

```go
type FormatterFormat int

const (
    FormatterFormatPlain FormatterFormat = C.GHOSTTY_FORMATTER_FORMAT_PLAIN
    FormatterFormatVT    FormatterFormat = C.GHOSTTY_FORMATTER_FORMAT_VT
    FormatterFormatHTML   FormatterFormat = C.GHOSTTY_FORMATTER_FORMAT_HTML
)
```

| Format | Description |
|--------|-------------|
| `FormatterFormatPlain` | Plain text (no escape sequences) |
| `FormatterFormatVT` | VT sequences preserving colors, styles, URLs, etc. |
| `FormatterFormatHTML` | HTML with inline styles |

## FormatterOption Functions

| Option | Signature | Description |
|--------|-----------|-------------|
| `WithFormatterFormat` | `(f FormatterFormat) FormatterOption` | Output format (default: Plain) |
| `WithFormatterUnwrap` | `(unwrap bool) FormatterOption` | Unwrap soft-wrapped lines |
| `WithFormatterTrim` | `(trim bool) FormatterOption` | Trim trailing whitespace on non-blank lines |
| `WithFormatterExtraPalette` | `(v bool) FormatterOption` | Emit palette via OSC 4 |
| `WithFormatterExtraModes` | `(v bool) FormatterOption` | Emit modes via CSI h/l |
| `WithFormatterExtraScrollingRegion` | `(v bool) FormatterOption` | Emit scrolling region (DECSTBM/DECSLRM) |
| `WithFormatterExtraTabstops` | `(v bool) FormatterOption` | Emit tabstop positions |
| `WithFormatterExtraPwd` | `(v bool) FormatterOption` | Emit pwd via OSC 7 |
| `WithFormatterExtraKeyboard` | `(v bool) FormatterOption` | Emit keyboard modes (modifyOtherKeys) |
| `WithFormatterExtraCursor` | `(v bool) FormatterOption` | Emit cursor position via CUP |
| `WithFormatterExtraStyle` | `(v bool) FormatterOption` | Emit current SGR style |
| `WithFormatterExtraHyperlink` | `(v bool) FormatterOption` | Emit hyperlink via OSC 8 |
| `WithFormatterExtraProtection` | `(v bool) FormatterOption` | Emit protection mode (DECSCA) |
| `WithFormatterExtraKittyKeyboard` | `(v bool) FormatterOption` | Emit Kitty keyboard state |
| `WithFormatterExtraCharsets` | `(v bool) FormatterOption` | Emit charset designations |

## Common Patterns

### Extract Plain Text for Assertion

```go
f, _ := libghostty.NewFormatter(term,
    libghostty.WithFormatterFormat(libghostty.FormatterFormatPlain),
    libghostty.WithFormatterTrim(true),
    libghostty.WithFormatterUnwrap(true),
)
defer f.Close()

content, _ := f.FormatString()
if !strings.Contains(content, "expected output") {
    t.Errorf("unexpected: %q", content)
}
```

### Preserve VT Formatting for Replay

```go
f, _ := libghostty.NewFormatter(term,
    libghostty.WithFormatterFormat(libghostty.FormatterFormatVT),
    libghostty.WithFormatterExtraCursor(true),
    libghostty.WithFormatterExtraStyle(true),
    libghostty.WithFormatterExtraPalette(true),
    libghostty.WithFormatterExtraModes(true),
)
defer f.Close()

vtSnapshot, _ := f.Format()
// vtSnapshot can be replayed into another terminal
```

### Generate HTML Snapshot

```go
f, _ := libghostty.NewFormatter(term,
    libghostty.WithFormatterFormat(libghostty.FormatterFormatHTML),
)
defer f.Close()

html, _ := f.FormatString()
os.WriteFile("snapshot.html", []byte(html), 0644)
```

### Golden File Comparison

```go
f, _ := libghostty.NewFormatter(term,
    libghostty.WithFormatterFormat(libghostty.FormatterFormatPlain),
    libghostty.WithFormatterTrim(true),
    libghostty.WithFormatterUnwrap(true),
)
defer f.Close()

got, _ := f.FormatString()
want, _ := os.ReadFile("testdata/golden.txt")
if got != string(want) {
    t.Errorf("output mismatch:\ngot:  %q\nwant: %q", got, string(want))
}
```
