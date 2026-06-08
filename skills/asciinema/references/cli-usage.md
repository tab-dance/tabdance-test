# CLI Usage — asciinema Command Line

In-depth reference for the asciinema CLI: recording, playback, and management.

## Installation

```sh
# Linux (apt)
sudo apt install asciinema

# macOS (Homebrew)
brew install asciinema

# pip
pip3 install asciinema

# Nix
nix-env -iA nixpkgs.asciinema
```

## Recording

```sh
# Start recording
asciinema rec

# Record with title
asciinema rec --title "My Demo"

# Record with idle time limit (compress pauses > 2s)
asciinema rec -i 2

# Record to file (no upload)
asciinema rec demo.cast

# Record with custom command
asciinema rec -c "bash --norc"

# Record with env vars
asciinema rec --env SHELL,TERM,PATH
```

### Recording Options

| Option | Description |
|--------|-------------|
| `--title <title>` | Set recording title |
| `-c <command>` | Run specific command instead of shell |
| `-i <seconds>` | Idle time limit (compress pauses) |
| `--env <vars>` | Comma-separated env vars to capture (default: SHELL,TERM) |
| `--raw` | Raw output (no timing information) |
| `--overwrite` | Overwrite existing file |
| `--append` | Append to existing recording |
| `--command <cmd>` | Same as `-c` |
| `--yes` / `-y` | Auto-accept upload confirmation |

### During Recording

- **Ctrl+\\** — Pause/resume recording (since v3.0)
- **Ctrl+D** or `exit` — End recording
- On exit, asciinema prompts to upload, save, or discard

## Playback

```sh
# Play a recording
asciinema play demo.cast

# Play at 2x speed
asciinema play -s 2 demo.cast

# Play with idle time limit
asciinema play -i 1 demo.cast

# Play from URL
asciinema play https://asciinema.org/a/12345.cast

# Stream from stdin
cat demo.cast | asciinema play -
```

### Playback Options

| Option | Description |
|--------|-------------|
| `-s <speed>` | Playback speed multiplier (default: 1) |
| `-i <seconds>` | Idle time limit |
| `--pause-on-marker` | Pause on markers (since v3.0) |

## Cat (Plain Text Output)

```sh
# Output plain text
asciinema cat demo.cast
```

Strips all escape sequences and outputs plain text. Useful for searching or grepping recording content.

## Upload

```sh
# Upload to asciinema.org
asciinema upload demo.cast
```

Recordings uploaded to asciinema.org get a unique URL. Private by default (since v3).

## Authentication

```sh
# Install install ID and token
asciinema auth
```

Opens a browser to link your local asciinema installation with your asciinema.org account.

## Configuration

Config file: `~/.config/asciinema/config.toml`

```toml
[api]
url = https://asciinema.org
token = <your-token>

[record]
command = /bin/bash
env = SHELL,TERM
idle_time_limit = 2
title = ""

[play]
speed = 1
idle_time_limit = -1
```

## v3 Features (asciinema CLI 3.x)

- **Pause/resume** during recording (Ctrl+\\)
- **Marker support** — add markers during recording for navigation
- **Line-by-line output** — `asciinema cat` outputs one line per terminal line
- **Private uploads by default**
- **Theme capture** — automatically records terminal color theme
- **Improved TTY detection**
