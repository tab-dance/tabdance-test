package golden

import (
	"fmt"
	"os"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
	libghostty "go.mitchellh.com/libghostty"
)

type VerifyResult struct {
	Match bool
	Diff  string
}

func (r VerifyResult) ToJSON() string {
	if r.Match {
		return `{"match":true}`
	}
	escaped := strings.ReplaceAll(r.Diff, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	escaped = strings.ReplaceAll(escaped, "\n", `\n`)
	return fmt.Sprintf(`{"match":false,"diff":"%s"}`, escaped)
}

func VerifyOutput(terminal *libghostty.Terminal, goldenPath string, format string, trim, unwrap bool) (*VerifyResult, error) {
	actual, err := getTerminalContent(terminal, format, trim, unwrap)
	if err != nil {
		return nil, err
	}

	goldenData, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &VerifyResult{Match: false, Diff: fmt.Sprintf("golden file %q does not exist", goldenPath)}, nil
		}
		return nil, fmt.Errorf("read golden file: %w", err)
	}
	golden := string(goldenData)

	if actual == golden {
		return &VerifyResult{Match: true}, nil
	}

	diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(golden),
		FromFile: goldenPath,
		B:        difflib.SplitLines(actual),
		ToFile:   "actual",
		Context:  3,
	})
	if err != nil {
		return nil, fmt.Errorf("generate diff: %w", err)
	}

	return &VerifyResult{Match: false, Diff: diff}, nil
}

func UpdateGolden(terminal *libghostty.Terminal, goldenPath string, format string, trim, unwrap bool) error {
	content, err := getTerminalContent(terminal, format, trim, unwrap)
	if err != nil {
		return err
	}

	if err := os.WriteFile(goldenPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write golden file: %w", err)
	}

	return nil
}

func getTerminalContent(terminal *libghostty.Terminal, format string, trim, unwrap bool) (string, error) {
	f, err := libghostty.NewFormatter(terminal,
		libghostty.WithFormatterFormat(parseFormat(format)),
		libghostty.WithFormatterTrim(trim),
		libghostty.WithFormatterUnwrap(unwrap),
	)
	if err != nil {
		return "", err
	}
	defer f.Close()

	content, err := f.FormatString()
	if err != nil {
		return "", err
	}
	return content, nil
}

func parseFormat(s string) libghostty.FormatterFormat {
	switch s {
	case "vt":
		return libghostty.FormatterFormatVT
	case "html":
		return libghostty.FormatterFormatHTML
	default:
		return libghostty.FormatterFormatPlain
	}
}
