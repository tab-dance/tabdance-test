# Input Injection — Sending Keys and Mouse Events

Reference for injecting keyboard and mouse input into headless terminals in tabdance-test.

## Key Injection

### KeyEncoder Pipeline

```
Agent request (key name + mods)
  → Parse key name to libghostty.Key
  → Parse mods to libghostty.Mods
  → Create KeyEvent
  → KeyEncoder.Encode(event)
  → Encoded bytes (escape sequence)
  → Write to PTY master (child's stdin)
```

### Key Name Parsing

The MCP `send_keys` tool accepts key names in W3C format:

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
| `F1` through `F12` | `KeyF1` through `KeyF12` |
| `Space` | `KeySpace` |

### Modifier Parsing

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
func (m *ManagedTerminal) SendKeys(keys []KeyInput) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    for _, k := range keys {
        if k.Text != "" {
            // Send raw text to PTY
            m.Ptmx.Write([]byte(k.Text))
            continue
        }

        event := libghostty.NewKeyEvent()
        defer event.Close()

        event.SetAction(libghostty.KeyActionPress)
        event.SetKey(parseKey(k.Key))
        event.SetMods(parseMods(k.Mods))

        seq, err := m.KeyEncoder.Encode(event)
        if err != nil { return err }
        if seq != nil {
            m.Ptmx.Write(seq)
        }

        // Also send release event for Kitty keyboard protocol
        if m.Terminal.KittyKeyboardFlags() != libghostty.KittyKeyDisabled {
            event.SetAction(libghostty.KeyActionRelease)
            seq, _ = m.KeyEncoder.Encode(event)
            if seq != nil {
                m.Ptmx.Write(seq)
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
func (m *ManagedTerminal) SendMouse(input MouseInput) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    event := libghostty.NewMouseEvent()
    defer event.Close()

    event.SetAction(parseMouseAction(input.Action))
    event.SetButton(parseMouseButton(input.Button))
    event.SetMods(parseMods(input.Mods))
    event.SetPosition(libghostty.MousePosition{X: input.X, Y: input.Y})

    seq, err := m.MouseEncoder.Encode(event)
    if err != nil { return err }
    if seq != nil {
        m.Ptmx.Write(seq)
    }
    return nil
}
```

### MouseEncoder Sync

```go
m.MouseEncoder.SetOptFromTerminal(m.Terminal)
```

Reads mouse tracking mode and output format from terminal state.

## Focus Events

```go
func (m *ManagedTerminal) SendFocus(event string) error {
    var fe libghostty.FocusEvent
    switch event {
    case "gained":
        fe = libghostty.FocusGained
    case "lost":
        fe = libghostty.FocusLost
    }

    seq, _ := libghostty.FocusEncode(fe)
    if seq != nil {
        m.Ptmx.Write(seq)
    }
    return nil
}
```

## Paste Events

```go
func (m *ManagedTerminal) SendPaste(text string) error {
    bracketed, _ := m.Terminal.ModeGet(libghostty.ModeBracketedPaste)
    encoded, _ := libghostty.PasteEncode([]byte(text), bracketed)
    m.Ptmx.Write(encoded)
    return nil
}
```

## Common Patterns

### Type a String

```go
func (m *ManagedTerminal) TypeString(s string) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    for _, r := range s {
        event := libghostty.NewKeyEvent()
        event.SetAction(libghostty.KeyActionPress)
        event.SetUTF8(string(r))
        seq, _ := m.KeyEncoder.Encode(event)
        if seq != nil {
            m.Ptmx.Write(seq)
        }
        event.Close()
    }
    return nil
}
```

### Press Ctrl+C

```go
event := libghostty.NewKeyEvent()
defer event.Close()
event.SetAction(libghostty.KeyActionPress)
event.SetKey(libghostty.KeyC)
event.SetMods(libghostty.ModCtrl)
seq, _ := enc.Encode(event)
ptmx.Write(seq)
```

### Navigate with Arrow Keys

```go
for i := 0; i < count; i++ {
    event := libghostty.NewKeyEvent()
    event.SetAction(libghostty.KeyActionPress)
    event.SetKey(libghostty.KeyArrowUp)
    seq, _ := enc.Encode(event)
    ptmx.Write(seq)
    event.Close()
}
```
