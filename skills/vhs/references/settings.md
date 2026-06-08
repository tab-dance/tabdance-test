# Settings — VHS Configuration Options

In-depth reference for all VHS `Set` settings.

## Overview

Settings configure the terminal and recording. All settings (except `TypingSpeed`) must appear at the top of the tape file, before any non-setting/non-output command.

## Terminal Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `Shell` | `bash` | Shell to use. Supported: `bash`, `zsh`, `fish`, `powershell`, `pwsh`, `cmd`, `nu`, `osh`, `xonsh` |
| `FontSize` | `22` | Font size in points |
| `FontFamily` | JetBrains Mono + fallbacks | Font family name |
| `LetterSpacing` | `1.0` | Letter spacing |
| `LineHeight` | `1.0` | Line height multiplier |

## Window Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `Width` | `1200` | Viewport width in pixels |
| `Height` | `600` | Viewport height in pixels |
| `Padding` | `60` | Terminal padding in pixels |
| `Margin` | `0` | Margin around terminal in pixels |
| `MarginFill` | Background color | Margin fill color (hex) |
| `WindowBar` | `""` (none) | Window bar style: `Colorful`, `ColorfulRight`, `Rings`, `RingsRight` |
| `WindowBarSize` | `30` | Window bar height in pixels |
| `BorderRadius` | `0` | Border radius in pixels |

## Recording Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `TypingSpeed` | `50ms` | Per-character typing delay. Can be overridden per-command with `@time` |
| `Framerate` | `50` | Capture framerate in FPS |
| `PlaybackSpeed` | `1.0` | Output playback speed multiplier |
| `LoopOffset` | `0` | GIF loop start offset (percentage: `50%`) |
| `CursorBlink` | `true` | Whether cursor blinks in recording |

## Theme Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `Theme` | Default dark | Theme name (e.g., `Catppuccin Frappe`) or JSON object |

### Theme as JSON

```tape
Set Theme {
  "name": "custom",
  "black": "#282c34",
  "red": "#e06c75",
  "green": "#98c379",
  "yellow": "#e5c07b",
  "blue": "#61afef",
  "magenta": "#c678dd",
  "cyan": "#56b6c2",
  "white": "#dcdfe4",
  "brightBlack": "#636d83",
  "brightRed": "#e06c75",
  "brightGreen": "#98c379",
  "brightYellow": "#e5c07b",
  "brightBlue": "#61afef",
  "brightMagenta": "#c678dd",
  "brightCyan": "#56b6c2",
  "brightWhite": "#dcdfe4",
  "background": "#282c34",
  "foreground": "#dcdfe4",
  "selectionBackground": "#474e5b",
  "cursor": "#a3b8cc"
}
```

## Wait Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `WaitTimeout` | `15s` | Default timeout for `Wait` commands |
| `WaitPattern` | `/>$/` | Default regex pattern for `Wait` |

## Shell-Specific Behavior

Each shell has custom prompt and no-rc flags:

| Shell | No-RC Flag | Prompt |
|-------|-----------|--------|
| `bash` | `--norc --noprofile` | Custom PS1 |
| `zsh` | `--no-rcs` | Custom PS1 |
| `fish` | `-C "function fish_greeting; end"` | Custom fish_prompt |
| `powershell`/`pwsh` | `-NoProfile` | Custom prompt function |
| `cmd` | — | Custom PROMPT |
| `nu` | `--no-config-file` | Custom env.PROMPT_COMMAND |
| `osh` | `--norc` | Custom PS1 |
| `xonsh` | `--no-rc` | Custom $PROMPT |
