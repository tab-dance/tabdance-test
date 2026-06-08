---
name: tabdance-test
description: >
  Use when developing or debugging tabdance-test — the MCP server for headless terminal testing
  using go-libghostty. Covers the MCP tool interface, headless terminal lifecycle, key/mouse
  injection, terminal inspection, asciinema recording, VHS tape playback, golden file verification,
  and Bubble Tea application testing. Triggers on: "tabdance-test", "headless terminal test",
  "MCP terminal", "terminal MCP server", "ghostty MCP", "libghostty MCP", "send keys",
  "retrieve view", "golden file", "asciinema recording", "VHS tape", "bubbletea test",
  "tabdance test", "integration test terminal".
user-invocable: true
---

# Tabdance-Test Development Reference

Comprehensive reference for developing [tabdance-test](.) — the MCP server that provides headless terminal testing via go-libghostty. An agent can send keys, resize the terminal, retrieve the current view, create asciinema recordings, play VHS tapes, and verify output against golden files. Designed for testing Bubble Tea applications (e.g., tabdance) using an AI agent.

## Architecture

```
AI Agent
  → MCP protocol (stdio)
    → tabdance-test MCP server
      → go-libghostty (headless VT terminal)
        → libghostty-vt (C VT emulation library)
      → PTY (child process running TUI app)
      → Asciicast recorder
      → VHS tape parser + executor
      → Golden file comparator
      → Wait/poll regex matcher
```

## Sub-Resources

Load the reference that matches your task. When in doubt, load multiple references.

|| Keywords | Reference |
|----------|----------|
| MCP tools, tool interface, create_terminal, send_keys, send_text, resize, get_view, get_content, screenshot, reset, list_terminals, close_terminal, MCP protocol, tool schema | [references/mcp-tools.md](references/mcp-tools.md) |
| terminal lifecycle, headless terminal, PTY, child process, go-libghostty, NewTerminal, VTWrite, Resize, effect callbacks, WritePty, bidirectional communication | [references/terminal-lifecycle.md](references/terminal-lifecycle.md) |
| key injection, send_keys, KeyEncoder, KeyEvent, Key, Mods, KeyAction, Kitty keyboard, mouse injection, MouseEncoder, MouseEvent, focus events, paste events, scroll | [references/input-injection.md](references/input-injection.md) |
| terminal inspection, get_view, get_content, Formatter, RenderState, GridRef, Cell, Row, Style, cursor position, screen content, plain text, VT, HTML | [references/terminal-inspection.md](references/terminal-inspection.md) |
| asciinema recording, .cast, AsciicastRecorder, incremental recording, timing, event stream, header, output/input/marker/resize events | [references/asciinema-recording.md](references/asciinema-recording.md) |
| VHS tape playback, .tape, tape parser, tape executor, Type, Enter, Ctrl, Sleep, Wait, Screenshot, converting tape to go-libghostty operations | [references/vhs-playback.md](references/vhs-playback.md) |
| golden file, verification, comparison, plain text, VT snapshot, HTML snapshot, diff, assertion, testdata, update golden | [references/golden-files.md](references/golden-files.md) |
| wait_for, poll, regex, screen match, line match, timeout, elapsed | [references/wait-poll.md](references/wait-poll.md) |
| bubbletea testing, tea.Program, headless, WithInput, WithOutput, WithoutRenderer, testing TUI, agent-driven test, tabdance integration test | [references/bubbletea-testing.md](references/bubbletea-testing.md) |

## Quick Guide

- **What MCP tools does tabdance-test expose?** → [references/mcp-tools.md](references/mcp-tools.md)
- **How does the headless terminal lifecycle work?** → [references/terminal-lifecycle.md](references/terminal-lifecycle.md)
- **How do I send keys/mouse to the terminal?** → [references/input-injection.md](references/input-injection.md)
- **How do I read the terminal screen?** → [references/terminal-inspection.md](references/terminal-inspection.md)
- **How do I create asciinema recordings?** → [references/asciinema-recording.md](references/asciinema-recording.md)
- **How do I play VHS tapes?** → [references/vhs-playback.md](references/vhs-playback.md)
- **How do I verify output against golden files?** → [references/golden-files.md](references/golden-files.md)
- **How do I wait for content to appear?** → [references/wait-poll.md](references/wait-poll.md)
- **How do I test Bubble Tea applications?** → [references/bubbletea-testing.md](references/bubbletea-testing.md)

## Cross-Project References

- For go-libghostty API details (Terminal, KeyEncoder, MouseEncoder, Formatter, RenderState, GridRef, Selection, etc.), use the **ghostty** skill.
- For asciicast v2 format specification and asciinema CLI, use the **asciinema** skill.
- For VHS tape DSL, parsing, and recording, use the **vhs** skill.
- For Bubble Tea framework internals (Model/Update/View, commands, messages, renderer, testing, composition), use the **bubbletea** skill.
- For tabdance application internals (select command, provider system, shell integration, UI model), use the **tabdance** skill.
