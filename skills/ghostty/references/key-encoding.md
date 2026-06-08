# Key Encoding — Keyboard Input to Escape Sequences

In-depth reference for `KeyEncoder` in go-libghostty: encoding keyboard events into terminal escape sequences.

## Overview

`KeyEncoder` converts `KeyEvent` structs into the byte sequences a terminal application expects as keyboard input. It supports legacy encoding, the Kitty Keyboard Protocol, modifyOtherKeys, and various DEC modes.

```go
enc, _ := libghostty.NewKeyEncoder()
defer enc.Close()

event := libghostty.NewKeyEvent()
defer event.Close()
event.SetAction(libghostty.KeyActionPress)
event.SetKey(libghostty.KeyA)
event.SetMods(libghostty.ModCtrl)

seq, _ := enc.Encode(event) // e.g., []byte("\x1b[97;5;1m") with Kitty protocol
```

## KeyEncoder

```go
type KeyEncoder struct { ptr C.GhosttyKeyEncoder }
```

Not safe for concurrent use.

### Lifecycle

```go
func NewKeyEncoder() (*KeyEncoder, error)
func (enc *KeyEncoder) Close()
```

### Configuration Methods

| Method | Signature | Description |
|--------|-----------|-------------|
| `SetOptBool` | `(opt KeyEncoderOption, val bool)` | Set a boolean option |
| `SetOptKittyFlags` | `(flags KittyKeyFlags)` | Set Kitty keyboard protocol flags |
| `SetOptOptionAsAlt` | `(val OptionAsAlt)` | Set macOS option-as-alt behavior |
| `SetOptFromTerminal` | `(t *Terminal)` | Read modes/flags from terminal |

### Encode

```go
func (enc *KeyEncoder) Encode(event *KeyEvent) ([]byte, error)
```

Returns the encoded escape sequence. Not all key events produce output (e.g., unmodified modifier keys); in that case, returns `(nil, nil)`.

Uses a 128-byte stack buffer first; on `GHOSTTY_OUT_OF_SPACE`, allocates a heap buffer and retries.

## KeyEncoderOption Constants

| Constant | Value Type | Description |
|----------|-----------|-------------|
| `KeyEncoderOptCursorKeyApplication` | `bool` | DEC mode 1: cursor key application mode |
| `KeyEncoderOptKeypadKeyApplication` | `bool` | DEC mode 66: keypad key application mode |
| `KeyEncoderOptIgnoreKeypadWithNumlock` | `bool` | DEC mode 1035: ignore keypad with numlock |
| `KeyEncoderOptAltEscPrefix` | `bool` | DEC mode 1036: alt sends escape prefix |
| `KeyEncoderOptModifyOtherKeysState2` | `bool` | xterm modifyOtherKeys mode 2 |
| `KeyEncoderOptKittyFlags` | `KittyKeyFlags` | Kitty keyboard protocol flags |
| `KeyEncoderOptMacOSOptionAsAlt` | `OptionAsAlt` | macOS option-as-alt setting |
| `KeyEncoderOptBackarrowKeyMode` | `bool` | Backarrow key mode (false=0x7f, true=0x08) |

## KittyKeyFlags

```go
type KittyKeyFlags uint8
```

| Constant | Description |
|----------|-------------|
| `KittyKeyDisabled` | Disables Kitty keyboard protocol |
| `KittyKeyDisambiguate` | Disambiguating escape codes |
| `KittyKeyReportEvents` | Report press and release events |
| `KittyKeyReportAlternates` | Report alternate key codes |
| `KittyKeyReportAll` | Report all key events |
| `KittyKeyReportAssociated` | Report associated text |
| `KittyKeyAll` | All flags enabled |

## OptionAsAlt

```go
type OptionAsAlt int
```

| Constant | Description |
|----------|-------------|
| `OptionAsAltFalse` | Option key is not treated as alt |
| `OptionAsAltTrue` | Option key is treated as alt |
| `OptionAsAltLeft` | Only left option key |
| `OptionAsAltRight` | Only right option key |

## KeyEvent

```go
type KeyEvent struct { ptr C.GhosttyKeyEvent }
```

### Lifecycle

```go
func NewKeyEvent() (*KeyEvent, error)
func (e *KeyEvent) Close()
```

### Setters

| Method | Signature | Description |
|--------|-----------|-------------|
| `SetAction` | `(action KeyAction)` | Press, release, or repeat |
| `SetKey` | `(key Key)` | Physical key (W3C standard) |
| `SetMods` | `(mods Mods)` | Active modifiers |
| `SetConsumedMods` | `(mods Mods)` | Modifiers consumed by OS |
| `SetComposing` | `(composing bool)` | Part of IME dead key sequence |
| `SetUTF8` | `(text string)` | Resulting UTF-8 text |
| `SetUnshiftedCodepoint` | `(cp rune)` | Unshifted codepoint |

### Getters

| Method | Return | Description |
|--------|--------|-------------|
| `Action()` | `KeyAction` | Press, release, or repeat |
| `Key()` | `Key` | Physical key |
| `Mods()` | `Mods` | Active modifiers |
| `ConsumedMods()` | `Mods` | Modifiers consumed by OS |
| `Composing()` | `bool` | IME composing state |
| `UTF8()` | `string` | Resulting text |
| `UnshiftedCodepoint()` | `rune` | Unshifted codepoint |

## KeyAction

```go
type KeyAction int

const (
    KeyActionPress   KeyAction = C.GHOSTTY_KEY_ACTION_PRESS
    KeyActionRelease KeyAction = C.GHOSTTY_KEY_ACTION_RELEASE
    KeyActionRepeat  KeyAction = C.GHOSTTY_KEY_ACTION_REPEAT
)
```

## Key (W3C UI Events Standard)

Physical keys follow the W3C UI Events specification. "KeyA" remains "KeyA" regardless of keyboard layout (QWERTY, AZERTY, etc.).

Common keys: `KeyA`–`KeyZ`, `Digit0`–`Digit9`, `Enter`, `Tab`, `Escape`, `Backspace`, `Delete`, `Insert`, `Home`, `End`, `PageUp`, `PageDown`, `ArrowUp/Down/Left/Right`, `F1`–`F12`, `Space`, `ShiftLeft/Right`, `ControlLeft/Right`, `AltLeft/Right`, `MetaLeft/Right`, `CapsLock`, `NumLock`, etc.

All `Key` constants implement `String()`, `encoding.TextMarshaler`, and `encoding.TextUnmarshaler`. Parse with top-level `ParseKey(s string) (Key, error)`.

## Mods

```go
type Mods uint16
```

| Constant | Description |
|----------|-------------|
| `ModShift` | Shift |
| `ModCtrl` | Control |
| `ModAlt` | Alt |
| `ModSuper` | Super/Command |
| `ModCaps` | Caps Lock |
| `ModNum` | Num Lock |
| `ModShiftLeft` | Left Shift |
| `ModShiftRight` | Right Shift |
| `ModCtrlLeft` | Left Control |
| `ModCtrlRight` | Right Control |
| `ModAltLeft` | Left Alt |
| `ModAltRight` | Right Alt |
| `ModSuperLeft` | Left Super |
| `ModSuperRight` | Right Super |

Implements `String()`, `TextMarshaler`, `TextUnmarshaler`. Parse with `ParseMods(s string) (Mods, error)`.

## SetOptFromTerminal

```go
func (enc *KeyEncoder) SetOptFromTerminal(t *Terminal)
```

Reads the terminal's current modes and flags and applies them to the encoder. Sets:
- Cursor key application mode
- Keypad mode
- Alt escape prefix
- modifyOtherKeys state
- Kitty keyboard protocol flags

**Note**: macOS option-as-alt cannot be determined from terminal state and is reset to `OptionAsAltFalse`. Use `SetOptOptionAsAlt` afterward if needed.

## Common Patterns

### Simulating Keyboard Input for Testing

```go
enc, _ := libghostty.NewKeyEncoder()
defer enc.Close()
enc.SetOptFromTerminal(term)

// Simulate Ctrl+C
event := libghostty.NewKeyEvent()
defer event.Close()
event.SetAction(libghostty.KeyActionPress)
event.SetKey(libghostty.KeyC)
event.SetMods(libghostty.ModCtrl)

seq, _ := enc.Encode(event)
// seq = []byte("\x03") in legacy mode
// seq = []byte("\x1b[99;5;1m") with Kitty protocol
```

### Full Input Pipeline for Headless Testing

```go
// 1. Create terminal
term, _ := libghostty.NewTerminal(libghostty.WithSize(80, 24))
defer term.Close()

// 2. Create key encoder synced to terminal state
enc, _ := libghostty.NewKeyEncoder()
defer enc.Close()
enc.SetOptFromTerminal(term)

// 3. Create PTY write handler to capture responses
var ptyOut bytes.Buffer
term.SetEffectWritePty(func(_ *libghostty.Terminal, data []byte) {
    ptyOut.Write(data)
})

// 4. Feed application output to terminal
term.VTWrite(appOutput)

// 5. Encode and send keyboard input
event := libghostty.NewKeyEvent()
defer event.Close()
event.SetAction(libghostty.KeyActionPress)
event.SetKey(libghostty.KeyEnter)
seq, _ := enc.Encode(event)

// In a real PTY session, seq would be written to the child's stdin
// For headless testing, the application would process seq
```
