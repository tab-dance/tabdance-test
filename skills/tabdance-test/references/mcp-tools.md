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
  "cols": 80,
  "rows": 24,
  "scrollback": 10000,
  "command": "optional command to run",
  "env": {"KEY": "VALUE"},
  "working_directory": "/optional/path"
}
```

**Output:**
```json
{
  "terminal_id": "term_abc123",
  "cols": 80,
  "rows": 24
}
```

#### `list_terminals`

List all active terminals.

**Output:**
```json
{
  "terminals": [
    {"id": "term_abc123", "cols": 80, "rows": 24, "command": "bash"}
  ]
}
```

#### `close_terminal`

Close and destroy a terminal.

**Input:**
```json
{
  "terminal_id": "term_abc123"
}
```

#### `reset_terminal`

Full terminal reset (RIS).

**Input:**
```json
{
  "terminal_id": "term_abc123"
}
```

### Input Injection

#### `send_keys`

Send key events to a terminal. Encodes keys using KeyEncoder synced to terminal state.

**Input:**
```json
{
  "terminal_id": "term_abc123",
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

**Key format**: W3C key names (`KeyA`, `Enter`, `ArrowUp`, `F1`, etc.) or literal characters. Mods: `ctrl`, `alt`, `shift`, `super` (combinable with `+`).

#### `send_text`

Send raw text to the terminal's PTY stdin (not VT-encoded — goes directly to the child process).

**Input:**
```json
{
  "terminal_id": "term_abc123",
  "text": "ls -la\n"
}
```

#### `send_mouse`

Send mouse events.

**Input:**
```json
{
  "terminal_id": "term_abc123",
  "action": "press",
  "button": "left",
  "x": 10,
  "y": 5,
  "mods": ""
}
```

**Actions**: `press`, `release`, `move`, `scroll_up`, `scroll_down`
**Buttons**: `left`, `middle`, `right`, `4`–`8`

#### `send_focus`

Send focus event.

**Input:**
```json
{
  "terminal_id": "term_abc123",
  "event": "gained"
}
```

#### `send_paste`

Send paste event (bracketed-paste-aware).

**Input:**
```json
{
  "terminal_id": "term_abc123",
  "text": "pasted content"
}
```

### Terminal Inspection

#### `get_view`

Get the current terminal view as plain text.

**Input:**
```json
{
  "terminal_id": "term_abc123",
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
  "terminal_id": "term_abc123",
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
  "terminal_id": "term_abc123"
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
  "terminal_id": "term_abc123"
}
```

**Output:**
```json
{
  "cols": 80,
  "rows": 24,
  "width_px": 640,
  "height_px": 384
}
```

### Resize

#### `resize`

Resize the terminal.

**Input:**
```json
{
  "terminal_id": "term_abc123",
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
  "terminal_id": "term_abc123",
  "output_path": "recording.cast",
  "idle_time_limit": 2.0
}
```

#### `stop_recording`

Stop recording and save the .cast file.

**Input:**
```json
{
  "terminal_id": "term_abc123"
}
```

#### `add_marker`

Add a marker to the current recording.

**Input:**
```json
{
  "terminal_id": "term_abc123",
  "label": "test step 1"
}
```

### VHS Tape Playback

#### `play_tape`

Execute a VHS tape file against a terminal.

**Input:**
```json
{
  "terminal_id": "term_abc123",
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
  "terminal_id": "term_abc123",
  "golden_path": "testdata/golden.txt",
  "format": "plain",
  "trim": true,
  "unwrap": true,
  "update": false
}
```

**Output:**
```json
{
  "match": true,
  "diff": ""
}
```

Or on mismatch:

```json
{
  "match": false,
  "diff": "--- golden\n+++ actual\n@@ -1 +1 @@\n-expected output\n+actual output\n"
}
```

#### `update_golden`

Update a golden file with current terminal content.

**Input:**
```json
{
  "terminal_id": "term_abc123",
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
  "terminal_id": "term_abc123",
  "pattern": "regex pattern",
  "scope": "screen",
  "timeout": 15
}
```

**Scopes**: `screen` (any line), `line` (current line)

**Output:**
```json
{
  "found": true,
  "elapsed_ms": 234
}
```

Or on timeout:

```json
{
  "found": false,
  "elapsed_ms": 15000
}
```
