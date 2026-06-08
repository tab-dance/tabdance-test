# Input Injection — Sending Keys and Mouse Events

Reference for injecting keyboard and mouse input into headless terminals in tabdance-test.

## Key Injection

### KeyEncoder Pipeline

```
Agent request (key name + mods)
  → libghostty.ParseKey(key) → libghostty.Key
  → libghostty.ParseMods(mods) → libghostty.Mods
  → Create KeyEvent
  → KeyEncoder.Encode(event)
  → Encoded bytes (escape sequence)
  → Write to PTY master (child's stdin)
```

### Key Name Parsing

The MCP `send_keys` tool accepts key names in W3C format, parsed by `libghostty.ParseKey`:

| Input | libghostty.Key |
|-------|---------------|
| `KeyA` through `KeyZ` | `KeyA` through `KeyZ` |
| `a` through `z` | `KeyA` through `KeyZ` (physical key) |
| `Enter` | `KeyEnter` |
| `Tab` | `KeyTab` |
| `Escape` | `KeyEscape` |
| `Backspace` | `KeyBackspace` |
| `Delete` | `KeyDelete` |
| `Insert` | `KeyInsert` |
| `Home` | `KeyHome` |
| `End` | `KeyEnd` |
| `PageUp` | `KeyPageUp` |
| `PageDown` | `KeyPageDown` |
| `ArrowUp/Down/Left/Right` | `KeyArrowUp/Down/Left/Right` |
| `F1` through `F25` | `KeyF1` through `KeyF25` |
| `Space` | `KeySpace` |

### Modifier Parsing

Parsed by `libghostty.ParseMods`:

| Input | libghostty.Mods |
|-------|----------------|
| `ctrl` | `ModCtrl` |
| `alt` | `ModAlt` |
| `shift` | `ModShift` |
| `super` | `ModSuper` |
| `ctrl+alt` | `ModCtrl | ModAlt` |
| `ctrl+shift+alt` | `ModCtrl | ModShift | ModAlt` |

### Implementation

```go
type KeyInput struct {
    Key  string `json:"key"`
    Mods string `json:"mods,omitempty"`
    Text string `json:"text,omitempty"`
}

func SendKeys(enc *libghostty.KeyEncoder, ptmxWriter io.Writer, term *libghostty.Terminal, keys []KeyInput) error {
    for _, k := range keys {
        if k.Text != "" {
            ptmxWriter.Write([]byte(k.Text))
            continue
        }

        key, _ := libghostty.ParseKey(k.Key)
        var mods libghostty.Mods
        if k.Mods != "" {
            mods, _ = libghostty.ParseMods(k.Mods)
        }

        event, _ := libghostty.NewKeyEvent()
        event.SetAction(libghostty.KeyActionPress)
        event.SetKey(key)
        event.SetMods(mods)

        seq, _ := enc.Encode(event)
        event.Close()
        if seq != nil {
            ptmxWriter.Write(seq)
        }

        // Also send release event for Kitty keyboard protocol
        kkf, _ := term.KittyKeyboardFlags()
        if kkf != libghostty.KittyKeyDisabled {
            releaseEvent, _ := libghostty.NewKeyEvent()
            releaseEvent.SetAction(libghostty.KeyActionRelease)
            releaseEvent.SetKey(key)
            releaseEvent.SetMods(mods)
            seq, _ = enc.Encode(releaseEvent)
            releaseEvent.Close()
            if seq != nil {
                ptmxWriter.Write(seq)
            }
        }
    }
    return nil
}
```

### Encoder Sync

After terminal state changes (resize, mode changes), re-sync the encoder:

```go
m.KeyEncoder.SetOptFromTerminal(m.Terminal)
```

This reads cursor key mode, keypad mode, alt-esc-prefix, modifyOtherKeys, and Kitty keyboard flags from the terminal.

## Mouse Injection

### MouseEncoder Pipeline

```
Agent request (action, button, position, mods)
  → Create MouseEvent
  → MouseEncoder.Encode(event)
  → Encoded bytes
  → Write to PTY master
```

### Implementation

```go
type MouseInput struct {
    Action string  `json:"action"`
    Button string  `json:"button"`
    X      float32 `json:"x"`
    Y      float32 `json:"y"`
    Mods   string  `json:"mods,omitempty"`
}

func SendMouse(enc *libghostty.MouseEncoder, ptmxWriter io.Writer, input MouseInput) error {
    event, _ := libghostty.NewMouseEvent()
    event.SetAction(parseMouseAction(input.Action))
    event.SetButton(libghostty.ParseMouseButton(input.Button))
    if input.Mods != "" {
        mods, _ := libghostty.ParseMods(input.Mods)
        event.SetMods(mods)
    }
    event.SetPosition(libghostty.MousePosition{X: input.X, Y: input.Y})

    seq, _ := enc.Encode(event)
    event.Close()
    if seq != nil {
        ptmxWriter.Write(seq)
    }
    return nil
}
```

### Scroll Support

Scroll actions (`scroll_up`/`scroll_down`) are translated to button 4/5 press+release:

```go
func SendScroll(enc *libghostty.MouseEncoder, ptmxWriter io.Writer, action string, x, y float32, mods string) error {
    // scroll_up → button 4, scroll_down → button 5
    // Sends both press and release events
}
```

### MouseEncoder Sync

```go
m.MouseEncoder.SetOptFromTerminal(m.Terminal)
```

Reads mouse tracking mode and output format from terminal state.

## Focus Events

```go
func SendFocus(ptmxWriter io.Writer, event string) error {
    fe, _ := libghostty.ParseFocusEvent(event)  // "gained" or "lost"
    seq, _ := libghostty.FocusEncode(fe)
    if seq != nil {
        ptmxWriter.Write(seq)
    }
    return nil
}
```

## Paste Events

```go
func SendPaste(ptmxWriter io.Writer, term *libghostty.Terminal, text string) error {
    bracketed, _ := term.ModeGet(libghostty.ModeBracketedPaste)
    encoded, _ := libghostty.PasteEncode([]byte(text), bracketed)
    if encoded != nil {
        ptmxWriter.Write(encoded)
    }
    return nil
}
```

## Common Patterns

### Type a String

The `Type` command in VHS tapes encodes each character individually:

```go
for _, r := range text {
    event, _ := libghostty.NewKeyEvent()
    event.SetAction(libghostty.KeyActionPress)
    event.SetUTF8(string(r))
    seq, _ := enc.Encode(event)
    event.Close()
    if seq != nil {
        ptmxWriter.Write(seq)
    }
    time.Sleep(typingSpeed)
}
```

### Press Ctrl+C

```go
key, _ := libghostty.ParseKey("KeyC")
mods, _ := libghostty.ParseMods("ctrl")
event, _ := libghostty.NewKeyEvent()
event.SetAction(libghostty.KeyActionPress)
event.SetKey(key)
event.SetMods(mods)
seq, _ := enc.Encode(event)
event.Close()
```
