# Bubble Tea Testing — Agent-Driven TUI Testing

Reference for testing Bubble Tea applications using tabdance-test's MCP server.

## Overview

tabdance-test enables AI agents to test Bubble Tea applications by:
1. Launching the app in a headless terminal
2. Sending key/mouse events
3. Reading the terminal view
4. Verifying output against golden files
5. Recording sessions as asciinema files

## Launching a Bubble Tea App

### Via PTY (Full Process)

```json
{
  "tool": "create_terminal",
  "arguments": {
    "cols": 108,
    "rows": 24,
    "command": "./tabdance select --out bash:/tmp/td-out"
  }
}
```

### Via Go Test (In-Process)

For in-process testing, use Bubble Tea's built-in test options:

```go
func TestTabdanceSelect(t *testing.T) {
    var buf bytes.Buffer
    var in bytes.Buffer

    m := newModel()
    p := tea.NewProgram(m,
        tea.WithInput(&in),
        tea.WithOutput(&buf),
        tea.WithWindowSize(108, 24),
    )

    go p.Run()

    // Send input
    in.Write([]byte("j\n"))  // navigate down, select

    // Wait for completion
    <-done

    // Verify output
    // ...
}
```

For Bubble Tea testing details, see the **bubbletea** skill → `references/testing.md`.

## Agent-Driven Test Flow

A typical agent-driven test session:

```
1. create_terminal → launch tabdance
2. wait_for /prompt/ → wait for app to render
3. get_view → verify initial state
4. send_keys ["j", "j", "Enter"] → navigate and select
5. wait_for /selected/ → wait for selection
6. get_view → verify selection result
7. verify_output → compare against golden file
8. stop_recording → save asciinema recording
9. close_terminal → cleanup
```

## Testing Tabdance Specifically

### Completion Menu

```json
// 1. Create terminal with tabdance
{"tool": "create_terminal", "arguments": {"cols": 108, "rows": 24, "command": "tabdance select --out bash:/tmp/td-out"}}

// 2. Wait for menu to render
{"tool": "wait_for", "arguments": {"terminal_id": "term_1", "pattern": "tabdance", "timeout": 5}}

// 3. Navigate down
{"tool": "send_keys", "arguments": {"terminal_id": "term_1", "keys": [{"key": "ArrowDown"}]}}

// 4. Select item
{"tool": "send_keys", "arguments": {"terminal_id": "term_1", "keys": [{"key": "Enter"}]}}

// 5. Verify output
{"tool": "verify_output", "arguments": {"terminal_id": "term_1", "golden_path": "/tmp/tabdance-test/testdata/golden/completion-plain.txt", "format": "plain"}}
```

### Context Preview

```json
// Navigate to trigger context loading
{"tool": "send_keys", "arguments": {"terminal_id": "term_1", "keys": [{"key": "ArrowDown"}, {"key": "ArrowDown"}]}}

// Wait for context to load
{"tool": "wait_for", "arguments": {"terminal_id": "term_1", "pattern": "git:", "scope": "screen", "timeout": 3}}

// Verify context pane content
{"tool": "get_view", "arguments": {"terminal_id": "term_1", "format": "plain", "trim": true}}
```

### Resize Testing

```json
// Resize to narrow terminal
{"tool": "resize", "arguments": {"terminal_id": "term_1", "cols": 54, "rows": 24}}

// Wait for reflow
{"tool": "wait_for", "arguments": {"terminal_id": "term_1", "pattern": "tabdance", "timeout": 2}}

// Verify layout adapted
{"tool": "verify_output", "arguments": {"terminal_id": "term_1", "golden_path": "/tmp/tabdance-test/testdata/golden/narrow-plain.txt", "format": "plain"}}
```

## Recording Test Sessions

```json
// Start recording before test
{"tool": "start_recording", "arguments": {"terminal_id": "term_1", "output_path": "/tmp/tabdance-test/recordings/completion-test.cast", "idle_time_limit": 2}}

// Add markers for test phases
{"tool": "add_marker", "arguments": {"terminal_id": "term_1", "label": "initial render"}}
// ... test steps ...
{"tool": "add_marker", "arguments": {"terminal_id": "term_1", "label": "after navigation"}}
// ... more steps ...
{"tool": "add_marker", "arguments": {"terminal_id": "term_1", "label": "after selection"}}

// Stop recording
{"tool": "stop_recording", "arguments": {"terminal_id": "term_1"}}
```

## VHS Tape for Tabdance

Create a `.tape` file for tabdance testing:

```tape
Output "tabdance-test.ascii"
Set Shell "bash"
Set Width 108
Set Height 24
Set TypingSpeed 100ms

Type "tabdance select --out bash:/tmp/td-out"
Enter
Wait /tabdance/
Sleep 1
ArrowDown
ArrowDown
Sleep 500ms
Enter
Wait /selected/
Sleep 1
```

Execute via MCP:

```json
{"tool": "play_tape", "arguments": {"terminal_id": "term_1", "tape_path": "/tmp/tabdance-test/tapes/tabdance.tape", "timeout": 30}}
```

## Common Assertions

### Check cursor position

```json
{"tool": "get_cursor", "arguments": {"terminal_id": "term_1"}}
// Verify cursor_x and cursor_y match expected values
```

### Check specific cell content

```json
{"tool": "get_cell", "arguments": {"terminal_id": "term_1", "x": 10, "y": 5}}
// Verify text, bold, fg_color match expected values
```

### Check terminal mode

Use `get_view` with VT format and check for mode-setting sequences, or use `wait_for` to detect mode-dependent behavior.

### Check for prompt (shell integration)

```json
{"tool": "wait_for", "arguments": {"terminal_id": "term_1", "pattern": "\\$", "timeout": 5}}
```

## Practical Usage Patterns

### Use a Temporary Directory for Test Artifacts

Always create a temp directory for recordings, golden files, and VHS tapes. This avoids polluting the project tree and makes cleanup trivial:

```json
// 1. Create temp directory (via bash before MCP calls)
// mkdir -p /tmp/tabdance-test/testdata/golden /tmp/tabdance-test/recordings

// 2. Use temp paths for all file outputs
{"tool": "start_recording", "arguments": {"terminal_id": "term_1", "output_path": "/tmp/tabdance-test/recordings/test.cast"}}
{"tool": "update_golden", "arguments": {"terminal_id": "term_1", "golden_path": "/tmp/tabdance-test/testdata/golden/initial.txt"}}
{"tool": "play_tape", "arguments": {"terminal_id": "term_1", "tape_path": "/tmp/tabdance-test/navigate.tape"}}
```

Only commit golden files to the project's `testdata/` when they're stable and meant for CI.

### Shell-Based Testing (Interactive Commands)

When testing a CLI tool that requires shell features (completion, pipes, env vars), launch `bash` first, then send commands via `send_text`:

```json
// 1. Create terminal with bash in the project directory
{"tool": "create_terminal", "arguments": {"cols": 108, "rows": 24, "command": "bash", "working_directory": "/tmp/tabdance-test"}}

// 2. Wait for shell prompt
{"tool": "wait_for", "arguments": {"terminal_id": "term_1", "pattern": "\\$|❯", "timeout": 5}}

// 3. Send command as text (not send_keys — send_text sends raw bytes to PTY)
{"tool": "send_text", "arguments": {"terminal_id": "term_1", "text": "tabdance select tabdance themes ''\n"}}

// 4. Wait for TUI to render
{"tool": "wait_for", "arguments": {"terminal_id": "term_1", "pattern": "SELECT", "timeout": 5}}

// 5. Navigate using send_keys (now works for special keys)
{"tool": "send_keys", "arguments": {"terminal_id": "term_1", "keys": [{"key": "ArrowDown"}, {"key": "ArrowDown"}]}}

// 6. Or use a small VHS tape for key sequences
// Write a .tape file to /tmp with: Down\nDown\nSleep 500ms
{"tool": "play_tape", "arguments": {"terminal_id": "term_1", "tape_path": "/tmp/tabdance-test/navigate.tape"}}
```

### send_text vs send_keys

- **`send_text`**: Sends raw bytes directly to PTY stdin. Use for printable text, commands, and `\n` for Enter. Does NOT go through KeyEncoder.
- **`send_keys`**: Encodes keys via KeyEncoder (respects terminal mode, Kitty keyboard protocol). Use for special keys: `ArrowUp`, `ArrowDown`, `Enter`, `Tab`, `Escape`, `F1`–`F12`, and modifier combos like `Ctrl+C`.
- **VHS tape `play_tape`**: Best for multi-step key sequences (e.g., `Down 3` then `Enter`). The tape executor handles key encoding correctly.

### Combining MCP Tools and VHS Tapes

For complex test flows, use MCP tools for waiting/verification and small VHS tapes for key input:

```
1. create_terminal (bash) + wait_for (prompt)
2. send_text (command + \n) + wait_for (TUI renders)
3. update_golden (to /tmp/.../testdata/golden/initial.txt)
4. play_tape (from /tmp/.../navigate.tape: Down, Down, Sleep)
5. get_view (verify navigation)
6. play_tape (from /tmp/.../select.tape: Enter, Sleep)
7. wait_for (result) + verify_output (against /tmp/.../golden)
8. stop_recording + close_terminal
```

### Common Pitfalls

- **`reset_terminal` clears VT state, not the child process**: After `reset_terminal`, the shell's readline buffer and prompt state are corrupted. Don't use `reset_terminal` to "start over" — instead `close_terminal` and `create_terminal` a fresh one.
- **`send_text` sends raw bytes, not escape sequences**: `send_text` with `"\x1b[B"` sends literal characters, not an arrow key. Use `send_keys` with `{"key": "ArrowDown"}` for special keys.
- **VHS `Screenshot` requires `.png` extension**: The upstream VHS parser rejects non-`.png` paths. In tabdance-test, screenshots are converted to markers anyway — use `add_marker` + `update_golden` instead.
- **`working_directory` sets the child process's cwd**: The terminal itself has no concept of a working directory — this only affects the child process started by `command`.
