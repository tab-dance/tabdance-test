# Golden Files — Output Verification

Reference for golden file verification in tabdance-test: comparing terminal output against expected content.

## Overview

Golden file testing compares the current terminal content against a stored "golden" file. If the content matches, the test passes. If not, the diff is shown for debugging.

## Golden File Formats

### Plain Text (Recommended)

Simplest and most readable. Strips all VT escape sequences.

```go
f, _ := libghostty.NewFormatter(term,
    libghostty.WithFormatterFormat(libghostty.FormatterFormatPlain),
    libghostty.WithFormatterTrim(true),
    libghostty.WithFormatterUnwrap(true),
)
defer f.Close()

got, _ := f.FormatString()
```

### VT Snapshot

Preserves colors, styles, and cursor position. Useful for verifying exact terminal state.

```go
f, _ := libghostty.NewFormatter(term,
    libghostty.WithFormatterFormat(libghostty.FormatterFormatVT),
    libghostty.WithFormatterExtraCursor(true),
    libghostty.WithFormatterExtraStyle(true),
    libghostty.WithFormatterExtraPalette(true),
)
defer f.Close()

got, _ := f.Format()
```

### HTML Snapshot

Useful for visual review in a browser.

```go
f, _ := libghostty.NewFormatter(term,
    libghostty.WithFormatterFormat(libghostty.FormatterFormatHTML),
)
defer f.Close()

got, _ := f.FormatString()
```

## Verification Flow

```go
func (m *ManagedTerminal) VerifyOutput(goldenPath, format string, trim, unwrap bool, update bool) (VerifyResult, error) {
    // 1. Get current terminal content
    got := m.GetViewContent(format, trim, unwrap)

    // 2. Read golden file
    want, err := os.ReadFile(goldenPath)
    if err != nil {
        if os.IsNotExist(err) && update {
            // Create new golden file
            os.WriteFile(goldenPath, []byte(got), 0644)
            return VerifyResult{Match: true}, nil
        }
        return VerifyResult{}, err
    }

    // 3. Compare
    if got == string(want) {
        return VerifyResult{Match: true}, nil
    }

    // 4. Generate diff
    diff := generateDiff(string(want), got)

    // 5. Update golden if requested
    if update {
        os.WriteFile(goldenPath, []byte(got), 0644)
    }

    return VerifyResult{
        Match: false,
        Diff:  diff,
    }, nil
}
```

## Diff Generation

```go
func generateDiff(want, got string) string {
    var buf bytes.Buffer
    diff := difflib.UnifiedDiff{
        A:        difflib.SplitLines(want),
        B:        difflib.SplitLines(got),
        FromFile: "golden",
        ToFile:   "actual",
        Context:  3,
    }
    difflib.WriteUnifiedDiff(&buf, diff)
    return buf.String()
}
```

## MCP Tool Usage

### Verify Output

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

Response on match:
```json
{"match": true, "diff": ""}
```

Response on mismatch:
```json
{
  "match": false,
  "diff": "--- golden\n+++ actual\n@@ -1,3 +1,3 @@\n-Hello, World!\n-$ ls\n+Hello, World!\n+$ ls -la\n"
}
```

### Update Golden

```json
{
  "terminal_id": "term_abc123",
  "golden_path": "testdata/golden.txt",
  "format": "plain",
  "trim": true,
  "unwrap": true
}
```

## Golden File Organization

```
testdata/
  golden/
    login-plain.txt
    login-vt.txt
    menu-selection-plain.txt
    menu-selection-vt.txt
  recordings/
    login.cast
    menu-selection.cast
  tapes/
    login.tape
    menu-selection.tape
```

## Best Practices

- Use `plain` format for most tests (readable, diffable)
- Use `vt` format when exact styling matters
- Always set `trim: true` and `unwrap: true` for consistent comparison
- Store golden files in git for version control
- Use `update: true` to create/update golden files, then review the diff before committing
- Use `wait_for` before verification to ensure the application has finished rendering
