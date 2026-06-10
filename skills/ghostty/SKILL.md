---
name: ghostty
description: >
  Use when working with Ghostty, libghostty, or go-libghostty — the terminal emulator,
  its C/Go VT emulation library, and the headless terminal API. Covers the Terminal type,
  VT write/resize, effect callbacks, key/mouse encoding, formatter, render state, grid
  references, selection system, Kitty graphics, configuration, shell integration, IPC,
  the apprt architecture, and building MCP servers or integration test harnesses on top
  of libghostty. Triggers on: "ghostty", "libghostty", "go-libghostty", "libghostty-vt",
  "GhosttyTerminal", "GhosttySurface", "ghostty config", "ghostty keybind",
  "ghostty shell-integration", "ghostty inspector", "ghostty IPC", "apprt",
  "headless terminal", "VT emulation", "NewTerminal", "VTWrite", "KeyEncoder",
  "MouseEncoder", "Formatter", "RenderState", "GridRef", "SelectionGesture",
  "KittyGraphics", "WritePtyFn", "BellFn", "ghostty.h", "ghostty/vt.h",
  "go.mitchellh.com/libghostty".
user-invocable: true
---

# Ghostty & libghostty In-Depth Reference

Comprehensive reference for [Ghostty](https://ghostty.org), [libghostty-vt](https://mitchellh.com/writing/libghostty-is-coming) (the standalone C VT emulation library), and [go-libghostty](https://tangled.org/mitchellh.com/go-libghostty) (the Go cgo bindings). Covers the headless Terminal API, VT parsing, input encoding, content formatting, render state, grid inspection, selection, Kitty graphics, configuration, shell integration, IPC, and the embedding/app runtime architecture.

## Sub-Resources

Load the reference that matches your task. When in doubt, load multiple references.

| Keywords | Reference |
|----------|----------|
| Terminal, NewTerminal, WithSize, VTWrite, Write, io.Writer, Resize, Reset, ModeGet, ModeSet, ScrollViewport, TerminalConfig, TerminalOption, TerminalScreen, Scrollbar, Close, lifecycle, headless terminal, creation, destruction | [references/terminal.md](references/terminal.md) |
| Effect, callback, WritePtyFn, BellFn, TitleChangedFn, EnquiryFn, XtversionFn, SizeFn, ColorSchemeFn, DeviceAttributesFn, WithWritePty, WithBell, WithTitleChanged, WithEnquiry, WithXtversion, WithSizeReport, WithColorScheme, WithDeviceAttributes, SetEffect, PTY response, query, OSC, CSI, XTWINOPS, DA, ENQ | [references/effects.md](references/effects.md) |
| KeyEncoder, KeyEvent, Key, KeyAction, Mods, KittyKeyFlags, OptionAsAlt, KeyEncoderOption, SetOptBool, SetOptFromTerminal, Encode, keyboard protocol, Kitty keyboard, modifyOtherKeys, cursor key application, keypad application, alt-esc-prefix, backarrow key mode, W3C UI Events | [references/key-encoding.md](references/key-encoding.md) |
| MouseEncoder, MouseEvent, MouseAction, MouseButton, MousePosition, MouseTrackingMode, MouseFormat, MouseEncoderSize, MouseEncoderOption, SetOptFromTerminal, Encode, X10, UTF-8, SGR, URxvt, SGR-Pixels, mouse tracking, motion dedup | [references/mouse-encoding.md](references/mouse-encoding.md) |
| FocusEvent, FocusEncode, PasteIsSafe, PasteEncode, bracketed paste, CSI I, CSI O, focus reporting | [references/focus-paste.md](references/focus-paste.md) |
| Formatter, NewFormatter, FormatterFormat, FormatterFormatPlain, FormatterFormatVT, FormatterFormatHTML, WithFormatterUnwrap, WithFormatterTrim, WithFormatterExtra*, Format, FormatString, WriteTo, io.WriterTo, content extraction, plain text, VT sequences, HTML | [references/formatter.md](references/formatter.md) |
| RenderState, NewRenderState, Update, RenderStateDirty, CursorVisualStyle, RenderStateColors, RenderStateRow, RenderStateCell, dirty tracking, incremental rendering, cell iteration, grapheme extraction, selection ranges | [references/render-state.md](references/render-state.md) |
| GridRef, TrackedGridRef, Point, PointTag, PointCoordinate, GridRef.Cell, GridRef.Row, GridRef.Graphemes, GridRef.HyperlinkURI, GridRef.Style, cell inspection, row inspection, coordinate systems, active, viewport, screen, history | [references/grid-reference.md](references/grid-reference.md) |
| Selection, SelectionGesture, SelectionGestureEvent, SelectWord, SelectLine, SelectAll, SelectOutput, SelectionFormat, SelectionAdjust, SelectionOrder, SelectionContains, SelectionEqual, word selection, line selection, output selection, rectangular selection, gesture state machine, click count, drag, autoscroll | [references/selection.md](references/selection.md) |
| KittyGraphics, KittyGraphicsImage, KittyGraphicsPlacementIterator, image storage, placement, layer, render info, RGB, RGBA, PNG, Zlib, virtual placement | [references/kitty-graphics.md](references/kitty-graphics.md) |
| Cell, Row, Style, StyleColor, ColorRGB, Palette, Mode, ModeReportState, DeviceAttributes, SizeReportSize, BuildInfo, Result, Error, screen types, style attributes, color system, terminal modes | [references/data-types.md](references/data-types.md) |
| ghostty config, config.ghostty, key = value, CLI flags, config-file, config-default-files, reloading, DerivedConfig, ConditionalState, RepeatableString, RepeatablePath, RepeatableStringMap, Keybinds, theme, palette, font, window, shell, env | [references/configuration.md](references/configuration.md) |
| keybind, trigger, action, key sequence, leader key, key table, key remap, consumed, all, global, performable, unconsumed, catch_all, chain, activate_key_table, W3C key codes, physical keys, unicode codepoints, modifier | [references/keybinds.md](references/keybinds.md) |
| shell-integration, shell-integration-features, cursor, sudo, title, ssh-env, ssh-terminfo, path, OSC 133, semantic prompt, prompt detection, command output | [references/shell-integration.md](references/shell-integration.md) |
| ghostty IPC, +new-window, D-Bus, Unix domain socket, gtk-single-instance, systemd, +list-actions, +show-config, +validate-config, +list-themes, +list-fonts, +list-keybinds, CLI actions, apprt, runtime, Surface, App, mailbox, action dispatch | [references/ipc-cli.md](references/ipc-cli.md) |
| ghostty.h, embedding API, ghostty_app_t, ghostty_surface_t, ghostty_config_t, ghostty_inspector_t, apprt, none.zig, GTK, macOS, SwiftUI, Metal, ImGui, dcimgui, inspector widgets, platform view | [references/embedding.md](references/embedding.md) |
| asciicast v2, .cast, JSON Lines, header, event stream, output, input, marker, resize, timing, theme, VHS, tape, .tape, DSL, Type, Sleep, Wait, Ctrl, Alt, Screenshot, golden file, integration test, ttyd, go-rod, ffmpeg | [references/recording-testing.md](references/recording-testing.md) |
| concurrency, thread safety, borrowed view, copied value, cgo.Handle, trampoline, Alloc, Free, static linking, dynamic linking, cross-compilation, Zig C compiler, pkg-config, build tags, SIMD, WASM | [references/cgo-build.md](references/cgo-build.md) |

## Quick Guide

- **How do I create a headless terminal?** → [references/terminal.md](references/terminal.md)
- **How do I feed VT data to a terminal?** → [references/terminal.md](references/terminal.md)
- **How do I handle terminal→host callbacks (PTY writes, bell, title)?** → [references/effects.md](references/effects.md)
- **How do I encode keyboard input for a terminal?** → [references/key-encoding.md](references/key-encoding.md)
- **How do I encode mouse input for a terminal?** → [references/mouse-encoding.md](references/mouse-encoding.md)
- **How do I encode focus and paste events?** → [references/focus-paste.md](references/focus-paste.md)
- **How do I extract terminal content as plain text, VT, or HTML?** → [references/formatter.md](references/formatter.md)
- **How do I read the render state for high-performance rendering?** → [references/render-state.md](references/render-state.md)
- **How do I inspect individual cells and rows?** → [references/grid-reference.md](references/grid-reference.md)
- **How do I perform selection operations?** → [references/selection.md](references/selection.md)
- **How do I work with Kitty graphics protocol images?** → [references/kitty-graphics.md](references/kitty-graphics.md)
- **What are the core data types (Cell, Row, Style, Color, Mode)?** → [references/data-types.md](references/data-types.md)
- **How does ghostty configuration work?** → [references/configuration.md](references/configuration.md)
- **How do keybindings and key tables work?** → [references/keybinds.md](references/keybinds.md)
- **How does shell integration work?** → [references/shell-integration.md](references/shell-integration.md)
- **How do I control a running ghostty instance via IPC?** → [references/ipc-cli.md](references/ipc-cli.md)
- **How do I embed ghostty in my application?** → [references/embedding.md](references/embedding.md)
- **How do I test with go-libghostty (golden files, asciicast generation)?** → [references/recording-testing.md](references/recording-testing.md)
- **How does cgo linking work and what are the concurrency rules?** → [references/cgo-build.md](references/cgo-build.md)

## Cross-Project References

- For Bubble Tea framework internals (Model/Update/View, commands, messages, renderer, testing, composition), use the **bubbletea** skill.
- For carapace library internals (Action API, traverse engine, shell formatters, style system, export format, UID system), use the **carapace-dev** skill.
- For carapace user-facing topics (integrating into CLIs, writing specs, macros, setup, choices), use the **carapace** skill.
- For asciinema/asciicast v2 format and CLI, use the **asciinema** skill.
- For VHS tape DSL, parsing, and recording, use the **vhs** skill.
- For in-depth shell-specific knowledge (bash, zsh, fish, elvish, nushell, powershell, xonsh, oil, tcsh), use the corresponding shell skill.
