# Configuration — Ghostty Config System

In-depth reference for Ghostty's configuration system: file format, config keys, loading, and runtime updates.

## Config File Format

### File Location

| Platform | Path |
|----------|------|
| All (XDG) | `$XDG_CONFIG_HOME/ghostty/config.ghostty` |
| All (XDG, legacy) | `$XDG_CONFIG_HOME/ghostty/config` |
| macOS (additional) | `$HOME/Library/Application Support/com.mitchellh.ghostty/config.ghostty` |

If `XDG_CONFIG_HOME` is not set, defaults to `$HOME/.config`.

### Syntax

```ini
# Comments start with # (only on their own line)
# Blank lines are ignored
background = 282c34
foreground = ffffff

# Whitespace around = doesn't matter
keybind = ctrl+z=close_surface

# Empty values reset to default
font-family =

# Values can be quoted or unquoted (equivalent)
font-family = "JetBrains Mono"
font-family = JetBrains Mono
```

- Keys are **case-sensitive** — always lowercase
- Every config key is also a valid CLI flag: `ghostty --background=282c34`
- Multiple files: `config-file = path` to include (relative to containing file)
- Prefix with `?` for optional includes: `config-file = ?~/.ghostty-local`
- `config-file` entries are processed at the **end** of the current file
- `config-default-files = false` (CLI-only) skips loading default config paths

### Reloading

Config can be reloaded at runtime: **Ctrl+Shift+,** (Linux) or **Cmd+Shift+,** (macOS). Some options require full restart.

## Config Key Categories

### Font & Language

| Key | Description |
|-----|-------------|
| `font-family` | Primary font (repeatable for fallbacks) |
| `font-family-bold` | Bold font |
| `font-family-italic` | Italic font |
| `font-family-bold-italic` | Bold italic font |
| `font-size` | Font size in points |
| `font-feature` | OpenType features (repeatable) |
| `font-variation` | Font variation axis values (repeatable) |
| `font-codepoint-map` | Map codepoints to specific fonts |
| `font-thicken` | Thicken font rendering (boolean) |
| `adjust-cell-width` | Cell width adjustment (float) |
| `adjust-cell-height` | Cell height adjustment (float) |
| `grapheme-width-method` | `unicode` or `double-width-as-wide` |

### Theme & Colors

| Key | Description |
|-----|-------------|
| `theme` | Built-in theme name, `custom`, or light/dark pair |
| `background` | Background color (hex or X11 name) |
| `foreground` | Foreground color |
| `palette` | Override 256-color palette entries (`palette = 0=#d54e53`) |
| `cursor-color` | Cursor color |
| `cursor-style` | `block`, `bar`, `underline`, `block_hollow` |
| `cursor-style-blink` | `true`, `false`, `in`, `out` |
| `selection-foreground` | Selection text color |
| `selection-background` | Selection background color |
| `background-opacity` | Background opacity (0.0–1.0) |
| `background-blur` | Background blur (Linux: integer, macOS: boolean) |

### Window & Layout

| Key | Description |
|-----|-------------|
| `window-padding-x` | Horizontal padding (points) |
| `window-padding-y` | Vertical padding (points) |
| `window-height` | Initial height in grid cells |
| `window-width` | Initial width in grid cells |
| `window-decoration` | `none`, `auto`, `client`, `server` |
| `window-theme` | `auto`, `system`, `light`, `dark`, `ghostty` |
| `fullscreen` | Start fullscreen (boolean) |
| `maximize` | Start maximized (boolean) |

### Command & Shell

| Key | Description |
|-----|-------------|
| `command` | Command to run (all surfaces) |
| `initial-command` | Command for first surface only |
| `env` | Environment variables (repeatable, `KEY=VALUE`) |
| `input` | Data to send on startup (`raw:` or `path:`) |
| `term` | TERM env var (default: `xterm-ghostty`) |
| `shell-integration` | `none`, `detect`, `bash`, `elvish`, `fish`, `nushell`, `zsh` |
| `shell-integration-features` | `cursor`, `sudo`, `title`, `ssh-env`, `ssh-terminfo`, `path` |
| `working-directory` | Initial working directory |

### Scrollback & Display

| Key | Description |
|-----|-------------|
| `scrollback-limit` | Scrollback buffer size in lines (default 10000) |
| `scrollbar` | `true`/`false` or `left`/`right`/`hidden`/`visible` |

### Clipboard

| Key | Description |
|-----|-------------|
| `clipboard-read` | `ask`, `allow`, `deny` |
| `clipboard-write` | `ask`, `allow`, `deny` |
| `clipboard-paste-protection` | Confirm before pasting unsafe text (boolean) |
| `copy-on-select` | Auto-copy on select (boolean) |

### Quick Terminal

| Key | Description |
|-----|-------------|
| `quick-terminal-position` | `top`, `bottom`, `left`, `right`, `center` |
| `quick-terminal-size` | Size as `%` or `px` |
| `quick-terminal-autohide` | Auto-hide on focus loss (boolean) |

### Linux/GTK-Specific

| Key | Description |
|-----|-------------|
| `gtk-single-instance` | `true`, `false`, `detect` |
| `gtk-tabs-location` | `top`, `bottom`, `hidden` |
| `gtk-custom-css` | Custom CSS file (repeatable) |
| `linux-cgroup` | `never`, `always`, `single-instance` |
| `async-backend` | `auto`, `epoll`, `io_uring` |

### macOS-Specific

| Key | Description |
|-----|-------------|
| `macos-titlebar-style` | `native`, `transparent`, `tabs`, `hidden` |
| `macos-option-as-alt` | `true`, `false`, `left`, `right` |
| `macos-non-native-fullscreen` | `true`, `false`, `visible-menu`, `padded-notch` |

## Special Field Types

| Type | Purpose | Example |
|------|---------|---------|
| `RepeatableString` | Multiple values | `font-family` specified multiple times |
| `RepeatablePath` | Multiple file paths | Theme directories |
| `RepeatableStringMap` | Key-value pairs | `env` for environment variables |
| `Keybinds` | Keybinding definitions | Multi-key sequences with actions |

## DerivedConfig Pattern

Each Surface maintains its own `DerivedConfig` — a snapshot of global config providing:
- **Thread safety**: Avoids sharing Config pointer across threads
- **Isolation**: Per-surface overrides without affecting others
- **Runtime updates**: Create new DerivedConfig and swap atomically

## Conditional Configuration

`ConditionalState` manages config that changes based on runtime conditions, particularly **light/dark mode theming**. Both App and Surface track `config_conditional_state`.

## Runtime Updates

`App.updateConfig()` propagates changes:
1. Sends `change_config` message to all surfaces
2. Applies conditional state changes
3. Notifies apprt via `performAction`
