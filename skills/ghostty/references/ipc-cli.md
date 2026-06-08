# IPC & CLI — Inter-Process Communication and Command Line

In-depth reference for Ghostty's IPC mechanism, CLI subcommands, and how to control a running Ghostty instance.

## CLI Syntax

```
ghostty [+action] [options]
```

If no `+action` is specified, Ghostty launches as the terminal emulator. Every config key is also a valid CLI flag.

## CLI Subcommands

### Informational

| Command | Description |
|---------|-------------|
| `+help` / `--help` | Show help (use with any action for specific help) |
| `+version` / `--version` | Display version info |

### Inspection & Listing

| Command | Flags | Description |
|---------|-------|-------------|
| `+list-fonts` | `--family`, `--bold`, `--italic`, `--style` | List available fonts |
| `+list-keybinds` | `--default`, `--docs`, `--plain` | List keybinds |
| `+list-themes` | `--path`, `--plain`, `--color` | List/preview themes |
| `+list-colors` | `--plain` | List named RGB colors |
| `+list-actions` | `--docs` | List keybind actions |

### Configuration

| Command | Flags | Description |
|---------|-------|-------------|
| `+show-config` | `--default`, `--changes-only`, `--docs` | Show current config |
| `+edit-config` | — | Open config in `$VISUAL`/`$EDITOR` |
| `+validate-config` | `--config-file` | Validate config file |
| `+explain-config` | — | Explain a config key or action |

### Runtime / IPC

| Command | Flags | Description |
|---------|-------|-------------|
| `+new-window` | `--class`, `--command`, `--working-directory`, `--title`, `-e` | Open new window in running instance |
| `+crash-report` | — | List crash reports |

### Utility

| Command | Description |
|---------|-------------|
| `+show-face` | Show font face for a codepoint |
| `+ssh-cache` | Manage SSH terminfo cache |
| `+boo` | Display Ghostty website animation |

## IPC: +new-window

The primary IPC mechanism for controlling a running Ghostty instance.

### Linux (GTK) — D-Bus

- Ghostty registers on the D-Bus session bus as `com.mitchellh.ghostty` (or custom class)
- If already running, D-Bus sends a signal to create a new window (~20ms vs ~300ms cold start)
- If not running, D-Bus activation starts Ghostty via systemd

### macOS — Unix Domain Socket

- Uses a Unix domain socket for IPC
- New process connects to socket, sends args, receives ack, exits
- Implemented in `src/apprt/embedded/ipc/new_window.zig`

### Flags

| Flag | Description |
|------|-------------|
| `--class=<class>` | Connect to specific Ghostty instance (GTK app ID / D-Bus name) |
| `--command=<cmd>` | Command to execute in first surface |
| `--working-directory=<dir>` | Working directory (relative to CWD) |
| `--title=<title>` | Override title of first surface |
| `-e` | Alias for `--command` (all subsequent args become part of command) |

### Examples

```sh
ghostty +new-window
ghostty +new-window --title="REMOTE: server1" --command="ssh server1"
ghostty +new-window --working-directory=/tmp --command=vim
ghostty +new-window --class=com.mitchellh.ghostty-debug
```

## Systemd Integration (Linux)

### Service Files

```
$PREFIX/share/applications/com.mitchellh.ghostty.desktop
$PREFIX/share/dbus-1/services/com.mitchellh.ghostty.service
$PREFIX/share/systemd/user/app-com.mitchellh.ghostty.service
```

### Commands

```sh
# Start at login (background)
systemctl enable --user app-com.mitchellh.ghostty.service

# Reload config (SIGUSR2)
systemctl reload --user app-com.mitchellh.ghostty.service

# View logs
journalctl -a -f --user -u app-com.mitchellh.ghostty.service

# Stop
systemctl stop --user app-com.mitchellh.ghostty.service
```

### gtk-single-instance

| Value | Description |
|-------|-------------|
| `true` | Every new process creates a window in existing instance |
| `false` | Each process launches separately |
| `detect` (default) | `true` unless `TERM_PROGRAM` is set or CLI args exist |

## -e Flag Behavior

`ghostty -e fish --with --args` sets `initial-command` and automatically:
- Disables shell expansion
- Sets `gtk-single-instance=false`
- Sets `quit-after-last-window-closed=true` (no delay)
- Sets `shell-integration=detect` (if not `none`)

## Environment Variables

| Variable | Description |
|----------|-------------|
| `GHOSTTY_RESOURCES_DIR` | Override resources directory |
| `TERM_PROGRAM` | Used by `gtk-single-instance=detect` |
| `XDG_CONFIG_HOME` | Config file location (default: `~/.config`) |

## Window Manager Integration

`ghostty +new-window` is designed to be bound to WM keys:

```
# Hyprland
bind = SUPER, return, exec, ghostty +new-window

# i3/Sway
bindsym $mod+Return exec ghostty +new-window
```
