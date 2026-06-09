# Configuration System

Reference for freeze's configuration — the Config struct, JSON config files, embedded templates, and user config.

## Overview

Freeze uses a layered configuration system: CLI flags → JSON config file → embedded defaults. The `Config` struct is defined in `config.go` (package main, unexported) and parsed by Kong.

## Config Struct

```go
type Config struct {
    Input string

    Background string
    Margin     []float64
    Padding    []float64
    Window     bool
    Width      float64
    Height     float64

    Version     bool
    Config      string
    Interactive bool
    Language    string
    Theme       string
    Wrap        int

    Output         string
    Execute        string
    ExecuteTimeout time.Duration

    Border Border
    Shadow Shadow
    Font   Font

    LineHeight      float64
    Lines           []int
    ShowLineNumbers bool
}
```

### Sub-Structs

```go
type Shadow struct {
    Blur float64
    X    float64
    Y    float64
}

type Border struct {
    Radius float64
    Width  float64
    Color  string
}

type Font struct {
    Family    string
    File      string
    Size      float64
    Ligatures bool
}
```

## Padding and Margin Expansion

Both `Padding` and `Margin` accept 1, 2, or 4 values (CSS-style shorthand):

| Input Length | Expansion (top, right, bottom, left) |
|-------------|--------------------------------------|
| 1 value | all four sides = value × scale |
| 2 values | top/bottom = first × scale, left/right = second × scale |
| 4 values | each side independently × scale |
| 0 or other | all zeros |

```go
func expandPadding(p []float64, scale float64) []float64
var expandMargin = expandPadding
```

The `scale` parameter is `1` for SVG output and `4` for auto-sized PNG output (to produce high-resolution images).

## Embedded Configurations

Freeze embeds JSON config files at compile time:

```go
//go:embed configurations/*
var configs embed.FS
```

### `base.json` — Minimal Config

```json
{
  "window": false,
  "theme": "charm",
  "border": { "radius": 0, "width": 0, "color": "#515151" },
  "shadow": { "blur": 0, "x": 0, "y": 0 },
  "padding": [20, 40, 20, 20],
  "margin": "0",
  "background": "#171717",
  "font": { "family": "JetBrains Mono", "size": 14, "ligatures": true },
  "line_height": 1.2
}
```

### `full.json` — Full Config

```json
{
  "window": true,
  "theme": "charm",
  "border": { "radius": 8, "width": 1, "color": "#515151" },
  "shadow": { "blur": 24, "x": 0, "y": 12 },
  "padding": [20, 40, 20, 20],
  "margin": [50, 60, 70, 60],
  "background": "#171717",
  "font": { "family": "JetBrains Mono", "size": 14, "ligatures": true },
  "line_height": 1.2
}
```

## Config Resolution Order

1. `--config default` (default) → loads `configurations/default.json` from embedded FS
2. `--config base` → loads `configurations/base.json`
3. `--config full` → loads `configurations/full.json`
4. `--config user` → loads `~/.config/freeze/user.json`
5. `--config /path/to/file.json` → loads custom JSON file
6. If any config file fails to open → falls back to `configurations/base.json`

The JSON config is loaded as a Kong resolver, so it provides defaults that CLI flags can override.

## User Config

Path: `~/.config/freeze/user.json` (via `xdg.ConfigHome`).

When using `--interactive` mode with the default config, freeze saves the interactive choices to the user config file for future sessions. The `Input`, `Output`, and `Interactive` fields are cleared before saving.

```go
var userConfigPath = filepath.Join(xdg.ConfigHome, "freeze", "user.json")
```

## Side Constants

Padding/margin indices use a `side` type:

```go
type side int
const (
    top    side = 0
    right  side = 1
    bottom side = 2
    left   side = 3
)
```

## Default Values

| Field | Default |
|-------|---------|
| `Config` | `"default"` |
| `Theme` | `"charm"` |
| `Background` | `"#171717"` |
| `Font.Family` | `"JetBrains Mono"` |
| `Font.Size` | `14` |
| `Font.Ligatures` | `true` |
| `LineHeight` | `1.2` |
| `Wrap` | `0` |
| `Lines` | `[0, -1]` (all lines) |
| `Output` | `""` → `freeze.png` |
| `ExecuteTimeout` | `10s` |
