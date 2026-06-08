# MCP Tools — Tool Interface Specification

Reference for the MCP tools exposed by tabdance-test.

## Overview

tabdance-test is an MCP server that provides tools for headless terminal testing. An AI agent uses these tools to create terminals, send input, inspect state, and verify output.

## Tool List

### Terminal Management

#### `create_terminal`

Create a new headless terminal, optionally running a child process.

**Input:**
```json
{
  "cols": 108,
  "rows": 24,
  "scrollback": 10000,
  "command": "optional command to run",
  "env": {"KEY": "VALUE"},
  "working_directory": "/tmp/tabdance-test"
}
```

**Output:**
```json
{
  "terminal_id": "term_1",
  "cols": 108,
  "rows": 24
}
```

#### `list_terminals`

List all active terminals.

**Output:**
```json
{
  "terminals": [
    {"id": "term_1", "cols": 108, "rows": 24, "command": "bash"}
  ]
}
```

#### `close_terminal`

Close and destroy a terminal.

**Input:**
```json
{
  "terminal_id": "term_1"
}
```

#### `reset_terminal`

Full terminal reset (RIS).

**Input:**
```json
{
  "terminal_id": "term_1"
}
```

### Input Injection

#### `send_keys`

Send key events to a terminal. Encodes keys using KeyEncoder synced to terminal state.

**Input:**
```json
{
  "terminal_id": "term_1",
  "keys": [
    {"key": "Enter"},
    {"key": "KeyA", "mods": "ctrl"},
    {"key": "c", "mods": "ctrl+shift"},
    {"text": "hello world"},
    {"key": "ArrowUp"},
    {"key": "Tab"}
  ]
}
```

**Key format**: W3C PascalCase key names (`Tab`, `Enter`, `ArrowUp`, `F1`, `KeyA`, etc.) or libghostty snake_case (`tab`, `enter`, `arrow_up`, `f1`, `key_a`). Both formats are accepted — PascalCase is preferred. Mods: `ctrl`, `alt`, `shift`, `super` (combinable with `+`).

#### `send_text`

Send raw text to the terminal's PTY stdin (not VT-encoded — goes directly to the child process).

**Input:**
```json
{
  "terminal_id": "term_1",
  "text": "ls -la\n"
}
```

#### `send_mouse`

Send mouse events.

**Input:**
```json
{
  "terminal_id": "term_1",
  "action": "press",
  "button": "left",
  "x": 10,
  "y": 5,
  "mods": ""
}
```

**Actions**: `press`, `release`, `move`, `scroll_up`, `scroll_down`
**Buttons**: `left`, `middle`, `right`, `4`–`8`

Scroll actions (`scroll_up`/`scroll_down`) are translated to button 4/5 press+release sequences automatically.

#### `send_focus`

Send focus event.

**Input:**
```json
{
  "terminal_id": "term_1",
  "event": "gained"
}
```

**Events**: `gained`, `lost`

#### `send_paste`

Send paste event (bracketed-paste-aware).

**Input:**
```json
{
  "terminal_id": "term_1",
  "text": "pasted content"
}
```

### Terminal Inspection

#### `get_view`

Get the current terminal view as plain text, VT, or HTML.

**Input:**
```json
{
  "terminal_id": "term_1",
  "format": "plain",
  "trim": true,
  "unwrap": true
}
```

**Formats**: `plain`, `vt`, `html`

**Output:**
```json
{
  "content": "Hello, World!\n$ _",
  "cursor_x": 3,
  "cursor_y": 1,
  "cursor_visible": true,
  "cursor_style": "bar",
  "active_screen": "primary"
}
```

#### `get_cell`

Get a specific cell's content and style.

**Input:**
```json
{
  "terminal_id": "term_1",
  "x": 10,
  "y": 5
}
```

**Output:**
```json
{
  "codepoint": 72,
  "text": "H",
  "wide": false,
  "bold": true,
  "fg": "#ff0000",
  "bg": "#000000",
  "hyperlink": ""
}
```

#### `get_cursor`

Get cursor position and style.

**Input:**
```json
{
  "terminal_id": "term_1"
}
```

**Output:**
```json
{
  "x": 3,
  "y": 1,
  "visible": true,
  "style": "bar",
  "pending_wrap": false
}
```

#### `get_size`

Get terminal dimensions.

**Input:**
```json
{
  "terminal_id": "term_1"
}
```

**Output:**
```json
{
  "cols": 108,
  "rows": 24,
  "width_px": 864,
  "height_px": 384
}
```

### Resize

#### `resize`

Resize the terminal.

**Input:**
```json
{
  "terminal_id": "term_1",
  "cols": 120,
  "rows": 40
}
```

### Recording

#### `start_recording`

Start asciinema recording for a terminal.

**Input:**
```json
{
  "terminal_id": "term_1",
  "output_path": "recording.cast",
  "idle_time_limit": 2.0
}
```

#### `stop_recording`

Stop recording and save the .cast file.

**Input:**
```json
{
  "terminal_id": "term_1"
}
```

#### `add_marker`

Add a marker to the current recording.

**Input:**
```json
{
  "terminal_id": "term_1",
  "label": "test step 1"
}
```

### VHS Tape Playback

#### `play_tape`

Execute a VHS tape file against a terminal.

**Input:**
```json
{
  "terminal_id": "term_1",
  "tape_path": "test.tape",
  "timeout": 30
}
```

### Golden File Verification

#### `verify_output`

Verify terminal output against a golden file.

**Input:**
```json
{
  "terminal_id": "term_1",
  "golden_path": "testdata/golden.txt",
  "format": "plain",
  "trim": true,
  "unwrap": true
}
```

**Output on match:**
```json
{"match": true}
```

**Output on mismatch:**
```json
{
  "match": false,
  "diff": "--- golden\n+++ actual\n@@ -1 +1 @@\n-expected output\n+actual output\n"
}
```

#### `update_golden`

Update (or create) a golden file with current terminal content. Separate tool to prevent accidental overwrites.

**Input:**
```json
{
  "terminal_id": "term_1",
  "golden_path": "testdata/golden.txt",
  "format": "plain",
  "trim": true,
  "unwrap": true
}
```

### Wait

#### `wait_for`

Wait for a pattern to appear on screen.

**Input:**
```json
{
  "terminal_id": "term_1",
  "pattern": "regex pattern",
  "scope": "screen",
  "timeout": 15
}
```

**Scopes**: `screen` (any line), `line` (current line)

**Output on match:**
```json
{
  "found": true,
  "elapsed_ms": 234
}
```

**Output on timeout:**
```json
{
  "found": false,
  "elapsed_ms": 15000
}
```

## Common Pitfalls

- **Use a temp directory for test artifacts**: Create `/tmp/tabdance-test/` for recordings, golden files, and VHS tapes. Only commit golden files to the project's `testdata/` when they're stable and meant for CI.
- **`send_text` sends raw bytes, not VT-encoded keys**: `send_text` writes directly to PTY stdin. Sending `"\x1b[B"` sends literal characters, not an arrow key. Use `send_keys` with `{"key": "ArrowDown"}` for special keys.
- **`send_keys` vs `send_text`**: Use `send_keys` for special keys (arrows, Enter, Tab, Escape, F-keys, modifier combos). Use `send_text` for printable text and commands (faster, simpler). For multi-step key sequences, use a small VHS tape via `play_tape`.
- **`reset_terminal` clears VT state, not the child process**: After reset, the shell's readline buffer and prompt are corrupted. To start fresh, `close_terminal` and `create_terminal` a new one instead.
- **`working_directory` sets the child process's cwd**: Only affects the process started by `command`. The terminal itself has no working directory concept.
- **VHS `Screenshot` requires `.png` extension**: The upstream VHS parser rejects non-`.png` paths. In tabdance-test, screenshots become markers — use `add_marker` + `update_golden` instead.
