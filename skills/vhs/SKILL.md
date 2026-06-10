---
name: vhs
description: >
  Use when working with VHS (charmbracelet/vhs) — the terminal GIF recorder that writes
  terminal recordings as code. Covers the .tape DSL, tape parsing, recording, screenshot
  capture, golden file testing, and CI integration. Triggers on: "vhs", "vhs tape", ".tape",
  "tape file", "tape DSL", "vhs record", "charmbracelet/vhs", "terminal GIF", "terminal recording",
  "Type[@time]", "Sleep", "Wait /regex/", "Screenshot", "golden file", "vhs-action",
  "ttyd", "go-rod".
user-invocable: true
---

# VHS In-Depth Reference

Comprehensive reference for [VHS](https://github.com/charmbracelet/vhs) (charmbracelet/vhs) — the CLI tool that writes terminal GIFs as code. Covers the `.tape` DSL, internal architecture, tape parsing, recording, and integration testing.

## Sub-Resources

Load the reference that matches your task. When in doubt, load multiple references.

| Keywords | Reference |
|----------|----------|
| tape DSL, Type, Sleep, Wait, Ctrl, Alt, Enter, Backspace, Tab, Screenshot, Hide, Show, Output, Require, Set, Env, Source, Copy, Paste, ScrollUp, ScrollDown, command reference | [references/tape-dsl.md](references/tape-dsl.md) |
| Set Shell, Set FontSize, Set FontFamily, Set Width, Set Height, Set TypingSpeed, Set Theme, Set Framerate, Set PlaybackSpeed, Set WaitTimeout, all settings | [references/settings.md](references/settings.md) |
| architecture, ttyd, go-rod, ffmpeg, xterm.js, canvas, frame recording, rendering pipeline, GIF, MP4, WebM, screenshot, compositing | [references/architecture.md](references/architecture.md) |
| lexer, parser, token, Command, CommandType, parser package, lexer package, token package, programmatic parsing, import | [references/tape-parsing.md](references/tape-parsing.md) |
| testing, golden file, .ascii, .txt, Buffer, integration test, CI, vhs-action, diff, assertion | [references/testing.md](references/testing.md) |
| vhs record, record mode, PTY, creack/pty, EscapeSequences, inputToTape, recording from terminal | [references/recording.md](references/recording.md) |

## Quick Guide

- **How do I write a .tape file?** → [references/tape-dsl.md](references/tape-dsl.md)
- **What settings are available?** → [references/settings.md](references/settings.md)
- **How does VHS work internally?** → [references/architecture.md](references/architecture.md)
- **How do I parse .tape files in Go?** → [references/tape-parsing.md](references/tape-parsing.md)
- **How do I use VHS for integration testing?** → [references/testing.md](references/testing.md)
- **How do I record a real terminal session?** → [references/recording.md](references/recording.md)
- **How do I take screenshots?** → [references/tape-dsl.md](references/tape-dsl.md) (Screenshot command)
- **How do I wait for specific output?** → [references/tape-dsl.md](references/tape-dsl.md) (Wait command)

## Cross-Project References

- For asciinema/asciicast v2 format, use the **asciinema** skill.
- For Ghostty/libghostty headless terminal testing, use the **ghostty** skill.
- For the tabdance-test MCP server (headless terminal testing), use the **tabdance-test** skill.
- For Bubble Tea framework internals, use the **bubbletea** skill.
