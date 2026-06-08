# Shell Integration — Prompt Detection and Semantic Markup

In-depth reference for Ghostty's shell integration system: how it injects shell-specific hooks, the features it enables, and the semantic markup it produces.

## Overview

Ghostty's shell integration injects shell-specific configuration to enable features like cursor styling, title reporting, SSH terminfo propagation, and semantic prompt markup (OSC 133).

## Configuration

### shell-integration

| Value | Description |
|-------|-------------|
| `none` | No automatic injection |
| `detect` (default) | Detect shell based on filename |
| `bash` | Force bash integration |
| `elvish` | Force elvish integration |
| `fish` | Force fish integration |
| `nushell` | Force nushell integration |
| `zsh` | Force zsh integration |

### shell-integration-features

Comma-separated list. Use `no-` prefix to disable. `true`/`false` for all.

| Feature | Default | Description |
|---------|---------|-------------|
| `cursor` | enabled | Set cursor to bar at prompt |
| `sudo` | enabled | Set sudo wrapper to preserve terminfo |
| `title` | enabled | Set window title via shell integration |
| `ssh-env` | enabled | Convert TERM to `xterm-256color` on remote hosts, propagate COLORTERM/TERM_PROGRAM/TERM_PROGRAM_VERSION |
| `ssh-terminfo` | enabled | Auto-install Ghostty terminfo on remote hosts via `infocmp`/`tic` |
| `path` | enabled | Add Ghostty's binary dir to PATH |

## OSC 133 — Semantic Prompt Markers

Shell integration injects OSC 133 sequences to mark up the terminal output with semantic regions:

| Sequence | Meaning |
|----------|---------|
| `OSC 133 ; A ST` | Mark start of prompt |
| `OSC 133 ; B ST` | Mark start of command input |
| `OSC 133 ; C ST` | Mark start of command output |
| `OSC 133 ; D [;exitcode] ST` | Mark end of command output |

These markers enable:
- **Jump to prompt**: `jump_to_prompt:N` keybind action
- **Output selection**: `SelectOutput` in go-libghostty
- **Semantic cell content**: `Cell.Semantic()` returns `CellSemanticContent` values

### CellSemanticContent

| Value | Description |
|-------|-------------|
| `CellSemanticContentInput` | Cell is part of command input |
| `CellSemanticContentOutput` | Cell is part of command output |
| `CellSemanticContentPrompt` | Cell is part of the prompt |

### RowSemanticPrompt

| Value | Description |
|-------|-------------|
| `RowSemanticPromptNone` | No prompt state |
| `RowSemanticPromptPrompt` | Row contains a prompt |
| `RowSemanticPromptContinuation` | Row is a prompt continuation |

## SSH Integration

When `ssh-env` and `ssh-terminfo` are both enabled:

1. Ghostty converts `TERM` to `xterm-256color` on remote hosts
2. Propagates `COLORTERM`, `TERM_PROGRAM`, `TERM_PROGRAM_VERSION`
3. Auto-installs Ghostty terminfo on remote hosts via `infocmp | tic -x -`
4. Cache managed with `ghostty +ssh-cache` CLI command
5. If terminfo installation fails, falls back to `xterm-256color`

### SSH Cache Management

```sh
ghostty +ssh-cache --clear           # Clear all cached terminfo
ghostty +ssh-cache --add --host=example.com  # Add host to cache
ghostty +ssh-cache --remove --host=example.com  # Remove host
ghostty +ssh-cache --expire-days=30  # Set expiry
```

## Cursor Integration

When `cursor` feature is enabled, shell integration sets the cursor to a bar shape at the prompt. This uses DECSCUSR sequences to change the cursor style when entering and leaving the prompt.

## Title Integration

When `title` feature is enabled, shell integration sets the window title via OSC 0/2 sequences. The title typically shows the current directory and running command.

## Sudo Integration

When `sudo` feature is enabled, shell integration wraps `sudo` to preserve the `TERMINFO` environment variable, ensuring that Ghostty's terminfo is available even when running commands with elevated privileges.

## Integration with go-libghostty

In a headless terminal, shell integration markers are parsed automatically by the VT parser. You can:

1. **Read semantic content** via `Cell.Semantic()` and `Row.Semantic()`
2. **Select command output** via `term.SelectOutput()`
3. **Detect prompts** by checking for `RowSemanticPromptPrompt`
4. **Feed integration markers** manually if needed: `term.VTWrite([]byte("\x1b]133;A\x07"))`

## TERM Value

Ghostty sets `TERM=xterm-ghostty` by default. This can be overridden with the `term` config key. The `xterm-ghostty` terminfo is installed by Ghostty and supports all Ghostty features including Kitty keyboard protocol, SGR mouse, and true color.
