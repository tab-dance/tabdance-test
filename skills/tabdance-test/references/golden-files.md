# Golden Files — Output Verification

Reference for golden file verification in tt: comparing terminal output against expected content.

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
    libghostty.WithFormatterTrim(true),
    libghostty.WithFormatterUnwrap(true),
)
defer f.Close()

got, _ := f.FormatString()
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

`verify_output` reads the golden file and compares against current terminal content. If the golden file doesn't exist, it returns `match: false` with a message (CLI: exit 1) - it does NOT auto-create. A content mismatch returns a tool error whose text contains the unified diff. Use `update_golden` to create or update golden files. `run_tape` combines play + verify (or update) with a throwaway terminal in one call.

```go
func VerifyOutput(terminal *libghostty.Terminal, goldenPath, format string, trim, unwrap bool) (*VerifyResult, error) {
    actual, _ := getTerminalContent(terminal, format, trim, unwrap)

    goldenData, err := os.ReadFile(goldenPath)
    if err != nil {
        if os.IsNotExist(err) {
            return &VerifyResult{Match: false, Diff: "golden file does not exist"}, nil
        }
        return nil, err
    }

    if actual == string(goldenData) {
        return &VerifyResult{Match: true}, nil
    }

    diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
        A:        difflib.SplitLines(string(goldenData)),
        FromFile: goldenPath,
        B:        difflib.SplitLines(actual),
        ToFile:   "actual",
        Context:  3,
    })

    return &VerifyResult{Match: false, Diff: diff}, nil
}
```

## Diff Generation

Uses `github.com/pmezard/go-difflib` for unified diff output:

```go
diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
    A:        difflib.SplitLines(golden),
    FromFile: "golden",
    B:        difflib.SplitLines(actual),
    ToFile:   "actual",
    Context:  3,
})
```

## MCP Tool Usage

### Verify Output

```json
{
  "terminal_id": "term_1",
  "golden_path": "testdata/golden.txt",
  "format": "plain",
  "trim": true,
  "unwrap": true
}
```

Response on match:
```json
{"match": true}
```

Response on mismatch:
```json
{
  "match": false,
  "diff": "--- golden\n+++ actual\n@@ -1,3 +1,3 @@\n-Hello, World!\n-$ ls\n+Hello, World!\n+$ ls -la\n"
}
```

### Update Golden

Separate tool to prevent accidental overwrites. Creates or replaces the golden file with current terminal content.

```json
{
  "terminal_id": "term_1",
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
- Use `update_golden` to create/update golden files, then review the diff before committing
- Use `wait_for` before verification to ensure the application has finished rendering
