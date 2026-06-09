# CLI Interface

Reference for the freeze CLI — flags, input modes, and output behavior.

## Overview

Freeze is a `package main` CLI application using [alecthomas/kong](https://github.com/alecthomas/kong) for argument parsing. It has no importable Go API — all functionality is accessed via the CLI.

## Usage Patterns

```
freeze main.go [-o code.svg] [--flags]
freeze --language go [-o code.png] < main.go
freeze --execute "ls -la" [-o output.svg] [--flags]
freeze -i  (interactive TUI mode)
```

## Flags

### Settings

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--output` | `-o` | `string` | `""` (falls back to `freeze.png`) | Output file path. Extension determines format: `.svg`, `.png`. |
| `--language` | `-l` | `string` | auto-detected | Language for syntax highlighting. Overrides auto-detection. |
| `--theme` | `-t` | `string` | `charm` | Theme name from chroma styles registry. |
| `--config` | `-c` | `string` | `default` | Configuration template: `base`, `full`, `user`, or path to JSON file. |
| `--wrap` | `-w` | `int` | `0` | Wrap lines at this character width. `0` = no wrapping. |
| `--execute` | `-x` | `string` | `""` | Execute a command and capture its ANSI output. Sets language to `ansi`. |
| `--execute.timeout` | — | `duration` | `10s` | Timeout for `--execute` command. |
| `--interactive` | `-i` | — | — | Launch interactive TUI for configuration. |
| `--version` | `-v` | — | — | Print version and exit. |

### Window

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--background` | `-b` | `string` | `#171717` | Background fill color. |
| `--margin` | `-m` | `[]float64` | `0` | Window margin. Accepts 1, 2, or 4 values (CSS-style: top, right, bottom, left). |
| `--padding` | `-p` | `[]float64` | `0` | Code padding. Accepts 1, 2, or 4 values. |
| `--window` | — | `bool` | `false` | Display macOS-style window controls (red/yellow/green dots). |
| `--width` | `-W` | `float64` | auto | Fixed width of terminal window in px. `0` = auto-size. |
| `--height` | `-H` | `float64` | auto | Fixed height of terminal window in px. `0` = auto-size. |

### Border

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--border.radius` | `float64` | `0` | Corner radius of window. |
| `--border.width` | `float64` | `0` | Border thickness in px. |
| `--border.color` | `string` | `#000` | Border color. |

### Shadow

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--shadow.blur` | — | `float64` | `0` | Gaussian blur radius. |
| `--shadow.x` | — | `float64` | `0` | Horizontal offset. |
| `--shadow.y` | — | `float64` | `0` | Vertical offset. |

### Font

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--font.family` | `string` | `JetBrains Mono` | Font family name. |
| `--font.file` | `string` | `""` | Path to custom font file (TTF, WOFF, WOFF2). |
| `--font.size` | `float64` | `14` | Font size in px. |
| `--font.ligatures` | `bool` | `true` (negatable) | Enable font ligatures. |

### Line

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--line-height` | `float64` | `1.2` | Line height relative to font size (em). |
| `--lines` | `[]int` | `0,-1` | Line range to capture (start, end). 1-indexed. |
| `--show-line-numbers` | `bool` | `false` | Display line numbers. |

## Input Modes

### File Input

```sh
freeze main.go -o code.svg
```

Language is auto-detected from the filename via `chroma.Get(filename)`.

### Stdin Input

```sh
echo 'package main' | freeze --language go -o code.svg
```

When stdin is a pipe (not a TTY), freeze reads from stdin. Language is auto-detected from content via `chroma.Analyse(input)`, or set explicitly with `--language`.

### Command Execution

```sh
freeze --execute "ls -la" -o output.png
```

Spawns a PTY, runs the command, captures output. Language is automatically set to `ansi`. Uses `xpty.NewPty` with terminal size from `term.GetSize(stdout)` (defaults to 80×24).

## Output Behavior

| Condition | Behavior |
|-----------|----------|
| `--output` with `.svg` extension | Writes SVG XML to file |
| `--output` with `.png` extension | Converts SVG to PNG (tries `rsvg-convert`, falls back to `resvg-go`) |
| No `--output`, stdin is a TTY | Writes to `freeze.png` (default filename) |
| No `--output`, stdin is a pipe | Writes SVG to stdout |
| File input, no `--output`, stdout is a TTY | Writes to `<filename>.svg` |
| File input, no `--output`, stdout is a pipe | Writes SVG to stdout |

## PNG Scale Factor

When both width and height are auto (default) and output is `.png`, freeze applies a **4× scale factor** to produce high-resolution output. This affects padding, margin, font size, and image dimensions.

## Gotchas

- **WebP is not actually supported**: Despite the `--output` help text mentioning `.webp`, freeze v0.2.2 silently produces SVG when given a `.webp` extension.
- **`--lines` is 1-indexed**: Internally converted to 0-indexed (`config.Lines[i]--`).
- **Auto-detection can fail**: If the language can't be determined, freeze exits with "Language Unknown". Use `--language` explicitly.
- **`--execute` sets language to `ansi`**: This is hardcoded, not configurable.
- **Config resolution order**: `default` → embedded `configurations/default.json` → `user` → `~/.config/freeze/user.json` → custom file path → fallback to `base.json`.
