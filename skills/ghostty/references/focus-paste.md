# Focus & Paste — Focus Events and Bracketed Paste

In-depth reference for focus event encoding and paste safety/encoding in go-libghostty.

## Focus Events

### FocusEvent

```go
type FocusEvent int

const (
    FocusGained FocusEvent = C.GHOSTTY_FOCUS_GAINED
    FocusLost   FocusEvent = C.GHOSTTY_FOCUS_LOST
)
```

Implements `String()`, `TextMarshaler`, `TextUnmarshaler`. Parse with `ParseFocusEvent(s string) (FocusEvent, error)`.

### FocusEncode

```go
func FocusEncode(event FocusEvent) ([]byte, error)
```

Encodes a focus event into a terminal escape sequence:
- `FocusGained` → `CSI I` (`\x1b[I`)
- `FocusLost` → `CSI O` (`\x1b[O`)

Only emitted when focus reporting is enabled (DEC mode 1004).

```go
seq, _ := libghostty.FocusEncode(libghostty.FocusGained)
// seq = []byte("\x1b[I")
```

## Paste

### PasteIsSafe

```go
func PasteIsSafe(data []byte) bool
```

Conservative safety check for pasted data. Returns `true` if the data is safe to paste without user confirmation. Uses the same heuristics as Ghostty's `clipboard-paste-protection` feature.

```go
if libghostty.PasteIsSafe(data) {
    // Safe to paste directly
} else {
    // Should confirm with user first
}
```

### PasteEncode

```go
func PasteEncode(data []byte, bracketed bool) ([]byte, error)
```

Encodes data for terminal input, optionally wrapping in bracketed paste sequences:
- `bracketed = true` → wraps with `\x1b[200~` … `\x1b[201~`
- `bracketed = false` → returns data as-is

```go
// With bracketed paste mode enabled
encoded, _ := libghostty.PasteEncode([]byte("hello world"), true)
// encoded = []byte("\x1b[200~hello world\x1b[201~")
```

## Common Patterns

### Sending Focus Events to a Terminal

```go
// Check if focus reporting is enabled
enabled, _ := term.ModeGet(libghostty.ModeFocusEvent)
if enabled {
    seq, _ := libghostty.FocusEncode(libghostty.FocusGained)
    // Send seq to the application's stdin
}
```

### Safe Paste Handling

```go
func sendPaste(term *libghostty.Terminal, data []byte) {
    bracketed, _ := term.ModeGet(libghostty.ModeBracketedPaste)
    encoded, _ := libghostty.PasteEncode(data, bracketed)
    // Write encoded to the application's stdin
}
```
