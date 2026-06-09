# Programmatic Integration

Reference for using freeze programmatically from Go — since freeze has no importable API, the only option is shelling out to the CLI.

## Overview

Freeze is `package main` — all core logic (Config struct, SVG generation, PNG conversion, ANSI handling) is unexported. The only importable packages are `svg`, `input`, and `font` (see [go-packages.md](go-packages.md)), which are low-level utilities that don't provide a complete image generation workflow.

## Shelling Out to freeze

### Basic: Generate SVG from stdin

```go
cmd := exec.Command("freeze", "--language", "go", "--output", "out.svg")
cmd.Stdin = strings.NewReader(`package main; func main() { println("hi") }`)
var stderr bytes.Buffer
cmd.Stderr = &stderr
if err := cmd.Run(); err != nil {
    return fmt.Errorf("freeze failed: %w: %s", err, stderr.String())
}
```

### Generate PNG from stdin

```go
cmd := exec.Command("freeze", "--language", "ansi", "--output", "out.png")
cmd.Stdin = strings.NewReader(ansiContent)
if err := cmd.Run(); err != nil {
    return fmt.Errorf("freeze failed: %w", err)
}
```

### Capture Command Output as Image

```go
cmd := exec.Command("freeze", "--execute", "ls -la", "--output", "output.png")
if err := cmd.Run(); err != nil {
    return fmt.Errorf("freeze failed: %w", err)
}
```

### With Custom Configuration

```go
cmd := exec.Command("freeze",
    "--language", "go",
    "--theme", "dracula",
    "--window",
    "--border.radius", "8",
    "--padding", "20,40",
    "--background", "#1e1e2e",
    "--output", "out.svg",
)
cmd.Stdin = strings.NewReader(sourceCode)
```

### Read Output Back into Memory

For SVG (text output):

```go
var stdout bytes.Buffer
cmd := exec.Command("freeze", "--language", "go")
cmd.Stdin = strings.NewReader(sourceCode)
cmd.Stdout = &stdout
if err := cmd.Run(); err != nil {
    return err
}
svgContent := stdout.String()
```

For PNG (binary output), write to a temp file then read it:

```go
tmpFile, err := os.CreateTemp("", "freeze-*.png")
if err != nil {
    return err
}
tmpPath := tmpFile.Name()
tmpFile.Close()
defer os.Remove(tmpPath)

cmd := exec.Command("freeze", "--language", "ansi", "--output", tmpPath)
cmd.Stdin = strings.NewReader(ansiContent)
if err := cmd.Run(); err != nil {
    return err
}
pngData, err := os.ReadFile(tmpPath)
```

## Integration Patterns

### Pattern: Terminal Screenshot from VT Content

When integrating with a headless terminal (e.g., go-libghostty), pipe VT-formatted content to freeze with `--language ansi`:

```go
func renderFreeze(vtContent string, format string, outputPath string) (string, error) {
    freezePath, err := exec.LookPath("freeze")
    if err != nil {
        return "", fmt.Errorf("freeze not found in PATH: %w", err)
    }

    if outputPath == "" {
        tmpFile, err := os.CreateTemp("", "freeze-*."+format)
        if err != nil {
            return "", err
        }
        outputPath = tmpFile.Name()
        tmpFile.Close()
    }

    args := []string{"--language", "ansi", "--output", outputPath}
    cmd := exec.Command(freezePath, args...)
    cmd.Stdin = strings.NewReader(vtContent)

    var stderr bytes.Buffer
    cmd.Stderr = &stderr

    if err := cmd.Run(); err != nil {
        return "", fmt.Errorf("freeze failed: %w: %s", err, stderr.String())
    }

    if format == "svg" {
        data, err := os.ReadFile(outputPath)
        if err != nil {
            return "", err
        }
        return string(data), nil
    }
    return outputPath, nil
}
```

### Pattern: Check for freeze Availability

```go
func isFreezeAvailable() bool {
    _, err := exec.LookPath("freeze")
    return err == nil
}
```

## Common Flags for Programmatic Use

| Flag | Recommended Value | Reason |
|------|-------------------|--------|
| `--language` | `ansi` | For terminal output with colors |
| `--output` | Explicit path | Avoid relying on default filename behavior |
| `--theme` | Omit (default `charm`) | ANSI input uses its own colors |
| `--padding` | `20,40` | Reasonable default for readability |
| `--background` | `#171717` | Matches default dark background |

## Gotchas

- **No library API**: Freeze cannot be imported as a Go library. All integration must shell out.
- **Stdout vs file output**: When stdin is a pipe and no `--output` is set, freeze writes SVG to stdout. When stdout is a TTY, it writes to a file. Be explicit with `--output` to avoid ambiguity.
- **Error output**: Freeze writes styled error messages to stderr. Capture stderr for error reporting.
- **Exit codes**: Freeze exits with code 1 on errors, 0 on success.
- **Temp file cleanup**: When using temp files for PNG output, remember to clean up with `defer os.Remove()`.
