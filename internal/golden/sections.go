package golden

import (
	"fmt"
	"os"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
)

// SectionSeparator starts each golden section; the text after it names
// the checkpoint that produced the section (e.g. `Expect /re/` or
// `final`). The dash row is identical in length to the separator the
// VHS .ascii goldens used, keeping the two generations visually and
// line-count comparable.
var SectionPrefix = strings.Repeat("─", 80) + " "

// SaveSections writes the golden as one labeled section per checkpoint
// followed by the final screen state (last section, label "final").
func SaveSections(path string, labels []string, contents []string) error {
	if len(labels) != len(contents) {
		return fmt.Errorf("sections length mismatch: %d labels, %d contents", len(labels), len(contents))
	}
	var sb strings.Builder
	for i, label := range labels {
		sb.WriteString(SectionPrefix + label + "\n")
		sb.WriteString(strings.TrimRight(contents[i], "\n") + "\n")
	}
	return os.WriteFile(path, []byte(sb.String()), 0644)
}

// LoadSections parses a labeled-section golden file. Every section
// begins with a `──── <label>` line; its content is everything after
// that up to the next separator. The last section carries the label
// `final` (the tape's end state).
func LoadSections(path string) ([]string, []string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	var labels, contents []string
	open := false
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, SectionPrefix) {
			labels = append(labels, strings.TrimSpace(strings.TrimPrefix(line, SectionPrefix)))
			contents = append(contents, "")
			open = true
			continue
		}
		if !open {
			continue // tolerate blank lines before the first section
		}
		idx := len(contents) - 1
		if contents[idx] != "" {
			contents[idx] += "\n"
		}
		contents[idx] += line
	}
	if len(labels) == 0 {
		return nil, nil, fmt.Errorf("no sections in golden file")
	}
	// The file's final newline was counted as an extra empty line; drop it.
	for i := range contents {
		contents[i] = strings.TrimSuffix(contents[i], "\n")
	}
	return labels, contents, nil
}

// SectionsText renders sections back into the on-disk golden format,
// without writing to a file; used to diff replayed sections against the
// committed golden via the usual string diff.
func SectionsText(labels, contents []string) string {
	var sb strings.Builder
	for i, label := range labels {
		sb.WriteString(SectionPrefix + label + "\n")
		sb.WriteString(strings.TrimRight(contents[i], "\n") + "\n")
	}
	return sb.String()
}

// DiffText renders a unified diff between two text blobs with stable
// file labels, for golden mismatch diagnostics.
func DiffText(expected, actual, goldenPath string) (string, error) {
	diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(expected),
		FromFile: goldenPath,
		B:        difflib.SplitLines(actual),
		ToFile:   "actual",
		Context:  3,
	})
	if err != nil {
		return "", fmt.Errorf("generate diff: %w", err)
	}
	return diff, nil
}

// Collector receives one checkpoint per successful anchor; the run path
// assembles labels/contents pairs from successive calls plus the final
// capture.
type Collector struct {
	labels   []string
	contents []string
}

func (c *Collector) Record(label, content string) {
	c.labels = append(c.labels, label)
	c.contents = append(c.contents, strings.TrimRight(content, "\n")+"\n")
}

func (c *Collector) Labels() []string   { return c.labels }
func (c *Collector) Contents() []string { return c.contents }
