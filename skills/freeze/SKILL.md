---
name: freeze
description: >
  Use when working with charmbracelet/freeze — the terminal and code screenshot tool that generates
  SVG and PNG images from source code, ANSI output, or command execution. Covers the CLI interface,
  configuration system (Config struct, JSON configs, embedded templates), ANSI rendering pipeline,
  SVG generation via Chroma, PNG conversion (rsvg-convert and resvg-go), font embedding, theme system,
  and programmatic integration (shelling out from Go). Triggers on: "freeze", "charmbracelet/freeze",
  "freeze CLI", "code screenshot", "terminal screenshot", "SVG from code", "PNG from code",
  "freeze config", "freeze theme", "freeze font", "freeze --execute", "freeze --language",
  "freeze --output", "rsvg-convert", "resvg-go", "JetBrains Mono freeze", "freeze configurations",
  "freeze ANSI", "freeze padding", "freeze margin", "freeze shadow", "freeze border",
  "freeze window controls", "freeze line numbers", "freeze interactive".
user-invocable: true
---

# Freeze In-Depth Reference

Comprehensive reference for [charmbracelet/freeze](https://github.com/charmbracelet/freeze) — the CLI tool that generates beautiful SVG and PNG images of source code and terminal output. Covers the CLI surface, configuration pipeline, ANSI rendering, SVG/PNG generation, font embedding, and programmatic integration.

## Data Flow

```
Input (file, stdin, or --execute)
  → language detection / --language override
    → Chroma lexer tokenization (or ANSI passthrough)
      → Chroma SVG formatter → etree XML document
        → layout calculation (padding, margin, dimensions)
          → decoration (window controls, border, shadow, line numbers)
            → ANSI dispatcher (if ANSI input: inject tspans/rects)
              → output dispatch:
                  .svg → write XML to file/stdout
                  .png → libsvgConvert (rsvg-convert) or resvgConvert (resvg-go)
```

## Sub-Resources

Load the reference that matches your task. When in doubt, load multiple references.

||| Keywords | Reference |
|||----------|----------|
|| CLI, flags, --output, --language, --theme, --execute, --window, --padding, --margin, --background, --width, --height, --wrap, --lines, --show-line-numbers, --interactive, --config, stdin, input modes, kong | [references/cli.md](references/cli.md) |
|| Config, Config struct, JSON config, embedded configurations, base.json, full.json, user config, Shadow, Border, Font, padding expansion, margin expansion, XDG config path | [references/configuration.md](references/configuration.md) |
|| ANSI, ANSI rendering, dispatcher, CSI dispatch, SGR, tspan, rect, background fill, 256-color palette, ansiPalette, CsiDispatch, Print, Execute, wide characters | [references/ansi-rendering.md](references/ansi-rendering.md) |
|| SVG, Chroma, etree, SVG generation, formatter, tokenization, lexer, font embedding, fontOptions, WOFF2, TTF, JetBrains Mono, font family, ligatures | [references/svg-generation.md](references/svg-generation.md) |
|| PNG, rsvg-convert, resvg-go, libsvgConvert, resvgConvert, DPI, pixmap, font database, scale factor, PNG conversion fallback | [references/png-conversion.md](references/png-conversion.md) |
|| theme, chroma styles, charm theme, style.go, chroma.StyleEntries, theme registry, catppuccin, dracula, gruvbox, nord, tokyonight, rose-pine, monokai | [references/themes.md](references/themes.md) |
|| svg package, AddShadow, AddClipPath, AddCornerRadius, Move, AddOutline, NewWindowControls, SetDimensions, GetDimensions, input package, ReadFile, ReadInput, IsPipe, font package, JetBrainsMonoTTF, JetBrainsMonoNLTTF | [references/go-packages.md](references/go-packages.md) |
|| programmatic, Go integration, shell out, exec.Command, piping stdin, temp file, integration patterns, embedding freeze in tools | [references/programmatic.md](references/programmatic.md) |

## Quick Guide

- **What CLI flags does freeze support?** → [references/cli.md](references/cli.md)
- **How does configuration loading work?** → [references/configuration.md](references/configuration.md)
- **How does freeze render ANSI escape sequences?** → [references/ansi-rendering.md](references/ansi-rendering.md)
- **How is the SVG document constructed?** → [references/svg-generation.md](references/svg-generation.md)
- **How does PNG conversion work?** → [references/png-conversion.md](references/png-conversion.md)
- **What themes are available?** → [references/themes.md](references/themes.md)
- **What Go packages does freeze export?** → [references/go-packages.md](references/go-packages.md)
- **How do I use freeze programmatically from Go?** → [references/programmatic.md](references/programmatic.md)

## Cross-Project References

- For Chroma (syntax highlighting library used by freeze), see the [Chroma documentation](https://github.com/alecthomas/chroma).
- For resvg (SVG rendering library used for PNG conversion), see [resvg-go](https://github.com/nickalie/resvg-go).
- For the Charm ecosystem (lipgloss, bubbletea, huh), use the **bubbletea** skill.
