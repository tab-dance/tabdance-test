# Wait/Poll — Regex-Based Content Matching

Reference for the `wait_for` MCP tool in tabdance-test: polling terminal content until a regex pattern matches.

## Overview

The `wait_for` tool polls terminal content at 50ms intervals until a regex pattern matches or the timeout expires. This is essential for testing async applications where rendering takes time.

## WaitFor

```go
type WaitForOptions struct {
    Pattern      string
    Scope        string        // "screen" or "line"
    Timeout      time.Duration
    PollInterval time.Duration // default: 50ms
    Format       string        // "plain", "vt", "html"
    Trim         bool
    Unwrap       bool
}

type WaitForResult struct {
    Found     bool
    MatchLine string  // set when scope is "line" and found
    ElapsedMs int64
}

func WaitFor(terminal *libghostty.Terminal, opts WaitForOptions) (*WaitForResult, error) {
    re, _ := regexp.Compile(opts.Pattern)

    start := time.Now()
    deadline := start.Add(opts.Timeout)

    for time.Now().Before(deadline) {
        content, _ := getTerminalContent(terminal, opts.Format, opts.Trim, opts.Unwrap)

        switch opts.Scope {
        case "line", "+Line":
            cursorY, _ := terminal.CursorY()
            lines := strings.Split(content, "\n")
            if int(cursorY) < len(lines) && re.MatchString(lines[cursorY]) {
                elapsed := time.Since(start)
                return &WaitForResult{Found: true, MatchLine: lines[cursorY], ElapsedMs: elapsed.Milliseconds()}, nil
            }
        default:
            if re.MatchString(content) {
                elapsed := time.Since(start)
                return &WaitForResult{Found: true, ElapsedMs: elapsed.Milliseconds()}, nil
            }
        }

        time.Sleep(50 * time.Millisecond)
    }

    return &WaitForResult{Found: false, ElapsedMs: opts.Timeout.Milliseconds()}, nil
}
```

## MCP Tool Usage

### Wait for any screen content

```json
{
  "terminal_id": "term_1",
  "pattern": "prompt",
  "scope": "screen",
  "timeout": 10
}
```

Response:
```json
{"found": true, "elapsed_ms": 234}
```

### Wait for current line match

```json
{
  "terminal_id": "term_1",
  "pattern": "\\$",
  "scope": "line",
  "timeout": 5
}
```

### Wait with format options

```json
{
  "terminal_id": "term_1",
  "pattern": "error|warning",
  "scope": "screen",
  "timeout": 15,
  "format": "plain",
  "trim": true,
  "unwrap": true
}
```

## Scope Behavior

| Scope | Matches Against |
|-------|----------------|
| `screen` (default) | Full terminal content (all visible lines) |
| `line` | Only the line at the current cursor Y position |

## Common Patterns

### Wait for shell prompt

```json
{"tool": "wait_for", "arguments": {"terminal_id": "term_1", "pattern": "\\$", "timeout": 5}}
```

### Wait for TUI to render

```json
{"tool": "wait_for", "arguments": {"terminal_id": "term_1", "pattern": "tabdance", "timeout": 10}}
```

### Wait for error output

```json
{"tool": "wait_for", "arguments": {"terminal_id": "term_1", "pattern": "error:|fatal:|panic:", "timeout": 30}}
```

### Chain with verify_output

```json
// 1. Wait for content to stabilize
{"tool": "wait_for", "arguments": {"terminal_id": "term_1", "pattern": "done", "timeout": 10}}
// 2. Then verify against golden file
{"tool": "verify_output", "arguments": {"terminal_id": "term_1", "golden_path": "testdata/result.txt"}}
```
