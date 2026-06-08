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
    "cols": 80,
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
        tea.WithWindowSize(80, 24),
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
{"tool": "create_terminal", "arguments": {"cols": 80, "rows": 24, "command": "tabdance select --out bash:/tmp/td-out"}}

// 2. Wait for menu to render
{"tool": "wait_for", "arguments": {"terminal_id": "term_1", "pattern": "tabdance", "timeout": 5}}

// 3. Navigate down
{"tool": "send_keys", "arguments": {"terminal_id": "term_1", "keys": [{"key": "ArrowDown"}]}}

// 4. Select item
{"tool": "send_keys", "arguments": {"terminal_id": "term_1", "keys": [{"key": "Enter"}]}}

// 5. Verify output
{"tool": "verify_output", "arguments": {"terminal_id": "term_1", "golden_path": "testdata/completion-plain.txt", "format": "plain"}}
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
{"tool": "resize", "arguments": {"terminal_id": "term_1", "cols": 40, "rows": 24}}

// Wait for reflow
{"tool": "wait_for", "arguments": {"terminal_id": "term_1", "pattern": "tabdance", "timeout": 2}}

// Verify layout adapted
{"tool": "verify_output", "arguments": {"terminal_id": "term_1", "golden_path": "testdata/narrow-plain.txt", "format": "plain"}}
```

## Recording Test Sessions

```json
// Start recording before test
{"tool": "start_recording", "arguments": {"terminal_id": "term_1", "output_path": "recordings/completion-test.cast", "idle_time_limit": 2}}

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
{"tool": "play_tape", "arguments": {"terminal_id": "term_1", "tape_path": "tests/tabdance.tape", "timeout": 30}}
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
