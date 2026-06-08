# Testing — Golden File and CI Integration

In-depth reference for using VHS in integration testing and CI pipelines.

## Golden File Testing

When the output is `.txt` or `.ascii`, VHS saves the terminal buffer content (not images) after each command step.

### How It Works

The `Buffer()` method reads the terminal text content via xterm.js:

```go
func (v *VHS) Buffer() ([]string, error) {
    buf, _ := v.Page.Eval("() => Array(term.rows).fill(0).map((e, i) => term.buffer.active.getLine(i).translateToString().trimEnd())")
    // Returns array of line strings
}
```

The `SaveOutput()` method writes the text content to a file, separated by `──────` lines between each command step.

### Creating a Golden File

```tape
Output "golden.ascii"
Set Shell "bash"
Type "echo hello"
Enter
Wait /hello/
```

```sh
vhs test.tape
# Creates golden.ascii with terminal content after each step
```

### Comparing Against Golden

```sh
# Run VHS and compare output
vhs test.tape
diff golden.ascii expected.ascii
```

Or in a test script:

```sh
#!/bin/bash
vhs test.tape
if diff -q golden.ascii expected.ascii > /dev/null; then
    echo "PASS"
else
    echo "FAIL: output differs from golden file"
    diff golden.ascii expected.ascii
    exit 1
fi
```

## CI Integration

### vhs-action (GitHub Actions)

```yaml
- uses: charmbracelet/vhs-action@v1
  with:
    path: test.tape
```

The action:
1. Installs VHS, ttyd, and ffmpeg
2. Runs the tape file
3. Uploads the output as an artifact
4. Fails if the tape execution errors

### Full CI Workflow

```yaml
name: VHS Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: charmbracelet/vhs-action@v1
        with:
          path: tests/demo.tape

      - name: Compare golden file
        run: diff tests/demo.ascii tests/golden/demo.ascii
```

### Requirements

VHS requires these dependencies in CI:

| Dependency | Purpose |
|------------|---------|
| `vhs` | Tape execution |
| `ttyd` | Terminal multiplexer |
| `ffmpeg` | Video compositing |
| Chromium/Chrome | go-rod browser |

## Limitations for Testing

| Limitation | Description |
|------------|-------------|
| No Go library API | VHS is a CLI tool; cannot import as a library |
| Requires display | Needs Chromium (headless) + ttyd |
| Slow startup | Browser + ttyd startup takes seconds |
| No direct PTY access | Goes through xterm.js, not a real PTY |
| Text-only comparison | Golden files are plain text, no style/color checking |
| No programmatic input | Cannot send input from Go code directly |

## go-libghostty as a Testing Alternative

For Go-based integration testing, go-libghostty provides a faster, more composable alternative:

| Feature | VHS | go-libghostty |
|---------|-----|---------------|
| Speed | Seconds (browser startup) | Microseconds (in-process) |
| Dependencies | Chromium, ttyd, ffmpeg | None (static linking) |
| CI-friendly | Requires display server | No display needed |
| Input control | DSL only | Full Go API |
| Output inspection | Plain text buffer | Plain/VT/HTML formatter |
| Style checking | No | Yes (Style, ColorRGB) |
| Cell-level inspection | No | Yes (GridRef, Cell) |
| Composability | CLI + parser packages | Full Go library |
