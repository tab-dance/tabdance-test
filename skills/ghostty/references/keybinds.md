# Keybinds — Key Binding System

In-depth reference for Ghostty's keybinding system: trigger format, actions, key sequences, key tables, key remapping, and trigger prefixes.

## Trigger Format

```
keybind = trigger=action
```

### Trigger Syntax

- **`+`-separated** list of keys and modifiers: `ctrl+a`, `ctrl+shift+b`, `up`
- **Unicode codepoints**: `a` matches any press producing that codepoint (layout-dependent, case-insensitive)
- **Physical keys**: W3C key codes like `KeyA`, `key_a` (snake_case also supported). Physical keys have higher priority than codepoints.
- **Function keys**: `insert`, `up`, `f5`, etc. (per W3C spec)
- **`catch_all`**: Matches any key not otherwise bound. Can combine with modifiers: `ctrl+catch_all`

### Modifiers

- `shift`, `ctrl` (alias: `control`), `alt` (aliases: `opt`, `option`), `super` (aliases: `cmd`, `command`)
- Modifiers cannot repeat (`ctrl+ctrl+a` is invalid)
- Only a single key input allowed (`ctrl+a+b` is invalid)
- Modifiers and keys can be in any order

## Key Sequences (Leader Keys)

Use `>` to require a sequence:

```
keybind = ctrl+a>n=new_window
```

- No timeout — waits indefinitely for next key
- If an unbound key is pressed, the entire sequence is sent to the program
- Previously bound prefixes are overridden by new sequences
- Not allowed for `global:` or `all:`-prefixed triggers

## Actions

| Action | Description |
|--------|-------------|
| `ignore` | Do nothing, black-hole the input |
| `unbind` | Remove the binding, key passes through to child |
| `csi:text` | Send CSI sequence (e.g., `csi:A` = cursor up) |
| `esc:text` | Send escape sequence (e.g., `esc:d`) |
| `text:text` | Send string using Zig string literal syntax (e.g., `text:\x15`) |
| `reset` | Reset terminal state |
| `clear_screen` | Clear screen and scrollback |
| `copy_to_clipboard` | Copy selection to clipboard |
| `paste_from_clipboard` | Paste from clipboard |
| `new_window` | Open new window |
| `new_tab` | Open new tab |
| `new_split:direction` | Open split (right/down/left/up/auto) |
| `close_surface` | Close current surface |
| `toggle_fullscreen` | Toggle fullscreen |
| `toggle_maximize` | Toggle maximize |
| `scroll_to_top` | Scroll to top of scrollback |
| `scroll_to_bottom` | Scroll to bottom |
| `scroll_page_up` | Scroll up one page |
| `scroll_page_down` | Scroll down one page |
| `jump_to_prompt:N` | Jump to Nth prompt (requires shell integration) |
| `increase_font_size:N` | Increase font size by N points |
| `decrease_font_size:N` | Decrease font size by N points |
| `reset_font_size` | Reset to default font size |
| `toggle_quick_terminal` | Toggle quick terminal |
| `toggle_command_palette` | Toggle command palette |
| `inspector:toggle` | Toggle terminal inspector |
| `reload_config` | Reload configuration |
| `activate_key_table:name` | Activate a key table |
| `deactivate_key_table` | Deactivate current key table |

Full list: `ghostty +list-actions --docs`

## Trigger Prefixes

| Prefix | Effect | Since |
|--------|--------|-------|
| `all:` | Apply to all terminal surfaces, not just focused | 1.0.0 |
| `global:` | System-wide keybind (works even when unfocused). Implies `all:`. | 1.0.0 (macOS), 1.2.0 (GTK) |
| `unconsumed:` | Don't consume input — also send encoded value to program | 1.0.0 |
| `performable:` | Only consume input if action is performable | 1.1.0 |

Multiple prefixes can be combined: `global:unconsumed:ctrl+a=reload_config`

## Chained Actions (Since 1.3.0)

```
keybind = ctrl+a=new_window
keybind = chain=goto_split:left
keybind = chain=toggle_fullscreen
```

Each `chain=` appends an action to the most recently defined keybind.

## Key Tables (Since 1.3.0)

Modal keybinding sets using `<table>/<binding>` syntax:

```
keybind = copy_mode/ctrl+a=new_window
```

- Activate: `activate_key_table:<name>`
- Deactivate: `deactivate_key_table`
- One-shot: `activate_key_table_once:<name>` (auto-deactivates on non-catch-all binding)
- `<name>/` alone defines and clears a table
- Lookup proceeds from innermost table outward
- Key sequences and chain actions work within tables
- Prefixes like `global:` work within tables

## Key Remap

```
key-remap = ctrl=super
key-remap = left_control=right_alt
```

- One-way remap (not transitive)
- Affects both keybind matching and terminal input encoding
- Generic names (`ctrl`) match both sides; sided names (`left_ctrl`) match one side
- Does NOT change keyboard layout or key encodings outside Ghostty

## Special Values

- `keybind=clear` — Clear ALL keybindings including defaults
- `keybind=unbind` as action — Remove the binding, key passes through

## Default Keybinds

View with: `ghostty +list-keybinds --default`

Key defaults include:
- `ctrl+shift+c` = copy_to_clipboard
- `ctrl+shift+v` = paste_from_clipboard
- `ctrl+shift+t` = new_tab
- `ctrl+shift+w` = close_surface
- `ctrl+shift+enter` = new_split:auto
- `ctrl+shift+,` = reload_config
- `ctrl+shift+i` = inspector:toggle
- `ctrl+plus` = increase_font_size:1
- `ctrl+minus` = decrease_font_size:1
- `ctrl+zero` = reset_font_size
