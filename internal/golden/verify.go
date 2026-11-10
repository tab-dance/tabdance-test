package golden

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/tab-dance/tabdance-test/internal/inspect"
	libghostty "go.mitchellh.com/libghostty"
)

type VerifyResult struct {
	Match bool   `json:"match"`
	Diff  string `json:"diff,omitempty"`
}

func (r VerifyResult) ToJSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

func VerifyOutput(terminal *libghostty.Terminal, goldenPath string, format string, trim, unwrap bool) (*VerifyResult, error) {
	actual, err := inspect.RenderContent(terminal, format, trim, unwrap)
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

	// For SVG, compare only the embedded JSON grid (not the full SVG markup)
	if format == "svg" {
		actualJSON, err := inspect.ExtractSVGGridJSON(actual)
		if err != nil {
			return nil, fmt.Errorf("extract JSON from actual SVG: %w", err)
		}
		goldenJSON, err := inspect.ExtractSVGGridJSON(golden)
		if err != nil {
			return nil, fmt.Errorf("extract JSON from golden SVG: %w", err)
		}
		if actualJSON == goldenJSON {
			return &VerifyResult{Match: true}, nil
		}
		diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
			A:        difflib.SplitLines(goldenJSON),
			FromFile: goldenPath + " (cell-grid JSON)",
			B:        difflib.SplitLines(actualJSON),
			ToFile:   "actual (cell-grid JSON)",
			Context:  3,
		})
		if err != nil {
			return nil, fmt.Errorf("generate diff: %w", err)
		}
		return &VerifyResult{Match: false, Diff: diff}, nil
	}

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
	content, err := inspect.RenderContent(terminal, format, trim, unwrap)
	if err != nil {
		return err
	}

	if err := os.WriteFile(goldenPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write golden file: %w", err)
	}

	return nil
}

// Capture renders the current terminal content with the given formatter
// style — the primitive behind golden sections and single-screen verify.
func Capture(terminal *libghostty.Terminal, format string, trim, unwrap bool) (string, error) {
	return inspect.RenderContent(terminal, format, trim, unwrap)
}
