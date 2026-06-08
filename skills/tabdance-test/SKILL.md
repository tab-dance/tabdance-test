---
name: tabdance-test
description: >
  Use when developing or debugging tabdance-test — the CLI tool and MCP server for headless terminal
  testing using go-libghostty. Covers the CLI command interface (Cobra + carapace), daemon architecture,
  MCP tool interface, headless terminal lifecycle, key/mouse injection, terminal inspection, asciinema
  recording, VHS tape playback, golden file verification, and Bubble Tea application testing.
  Triggers on: "tabdance-test", "headless terminal test", "MCP terminal", "terminal MCP server",
  "ghostty MCP", "libghostty MCP", "send keys", "retrieve view", "golden file",
  "asciinema recording", "VHS tape", "bubbletea test", "tabdance test", "integration test terminal",
  "tabdance-test exec", "tabdance-test run", "tabdance-test view", "tabdance-test keys",
  "daemon socket", "terminal daemon".
user-invocable: true
---

# Tabdance-Test Development Reference

Comprehensive reference for developing [tabdance-test](.) — the CLI tool and MCP server that provides headless terminal testing via go-libghostty. Supports two interfaces:

1. **CLI** (Cobra + carapace completions) — for local development and CI testing
2. **MCP server** (stdio) — for AI agent–driven testing

Both interfaces provide the same capabilities: create terminals, send keys/mouse, inspect state, record sessions, play VHS tapes, and verify output against golden files.

## Architecture

### CLI Mode (daemon-backed)

```
CLI client (tabdance-test view, keys, etc.)
  → Unix socket (JSON-RPC)
    → Daemon process (auto-forked, holds TerminalManager)
      → go-libghostty (headless VT terminal)
        → libghostty-vt (C VT emulation library)
      → PTY (child process running TUI app)
      → Asciicast recorder
      → VHS tape parser + executor
      → Golden file comparator
      → Wait/poll regex matcher
```

### CLI Mode (daemon-free)

```
tabdance-test run tape.tape --golden test.golden
  → in-process: create terminal → play tape → verify golden → exit 0/1
```

### MCP Mode

```
AI Agent
  → MCP protocol (stdio)
    → tabdance-test mcp
      → (same internal packages as daemon)
```

## Sub-Resources

Load the reference that matches your task. When in doubt, load multiple references.

|| Keywords | Reference |
|----------|----------|
| CLI commands, cobra, carapace, exec, view, keys, run, daemon, socket, attach, terminal ID completion | [references/cli-commands.md](references/cli-commands.md) |
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

- **What CLI commands does tabdance-test expose?** → [references/cli-commands.md](references/cli-commands.md)
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
- For carapace shell completion integration (cobra setup, ActionMap, positional completion, dynamic actions), use the **carapace** skill.
