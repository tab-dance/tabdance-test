package golden

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSectionRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tape.golden")
	labels := []string{`Wait+Screen /example "ac"/`, `Expect /action/`, "final"}
	contents := []string{
		"$ example \"ac\"",
		"$ example \"action\"",
		"$ example done",
	}
	if err := SaveSections(path, labels, contents); err != nil {
		t.Fatal(err)
	}
	gotLabels, gotContents, err := LoadSections(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotLabels) != len(labels) {
		t.Fatalf("loaded %d sections, want %d", len(gotLabels), len(labels))
	}
	for i := range labels {
		if gotLabels[i] != labels[i] {
			t.Errorf("section %d label = %q, want %q", i, gotLabels[i], labels[i])
		}
		if gotContents[i] != contents[i] {
			t.Errorf("section %d content = %q, want %q", i, gotContents[i], contents[i])
		}
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), strings.Repeat("─", 80)+" Wait+Screen /example \"ac\"/") {
		t.Errorf("separator prefix missing from file:\n%s", data)
	}
}

func TestSectionsTextMatchesSave(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tape.golden")
	labels := []string{`Expect /x/`, "final"}
	contents := []string{"one\n", "two\n"}
	if err := SaveSections(path, labels, contents); err != nil {
		t.Fatal(err)
	}
	saved, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if SectionsText(labels, contents) != string(saved) {
		t.Fatalf("SectionsText diverges from SaveSections:\n%s\nvs\n%s", SectionsText(labels, contents), saved)
	}
}

func TestLoadSectionsRejectsSectionlessFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tape.golden")
	if err := os.WriteFile(path, []byte("plain golden\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := LoadSections(path); err == nil {
		t.Fatal("expected error for file without sections")
	}
}
