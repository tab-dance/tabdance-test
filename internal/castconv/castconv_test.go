package castconv

import (
	"os"
	"strings"
	"testing"
)

func writeCast(t *testing.T, content string) string {
	t.Helper()
	path := t.TempDir() + "/rec.cast"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseBasicEvents(t *testing.T) {
	path := writeCast(t, `{"version": 2, "width": 80, "height": 24}
[0.1, "o", "hello\r\n"]
[0.2, "i", "echo hi\r"]
[1.0, "m", "after"]
[2.0, "r", "100x25"]
`)
	cast, err := Parse(path)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(cast.Events) != 4 {
		t.Fatalf("expected 4 events, got %d", len(cast.Events))
	}
	if cast.Width != 80 || cast.Height != 24 {
		t.Fatalf("header mismatch: %dx%d", cast.Width, cast.Height)
	}
	if cast.Events[1].Kind != "i" || cast.Events[1].Data != "echo hi\r" {
		t.Fatalf("input event lost: %+v", cast.Events[1])
	}
}

func TestDecodeInput(t *testing.T) {
	got := strings.Join(decodeInput("echo hello\r"), "\n")
	want := "Type \"echo hello\"\nEnter"
	if got != want {
		t.Fatalf("decodeInput =\n%s\nwant\n%s", got, want)
	}

	got = strings.Join(decodeInput("ls\x7f"), "\n")
	if !strings.Contains(got, "Backspace") || !strings.Contains(got, "Type \"ls\"") {
		t.Fatalf("backspace decoding wrong: %q", got)
	}

	got = strings.Join(decodeInput("\x03"), "\n")
	if got != "Ctrl+C" {
		t.Fatalf("ctrl decoding wrong: %q", got)
	}

	got = strings.Join(decodeInput("\x1b[B"), "\n")
	if got != "Down" {
		t.Fatalf("arrow decoding wrong: %q", got)
	}

	got = strings.Join(decodeInput("\x1bOP"), "\n")
	if got != "F1" {
		t.Fatalf("F key decoding wrong: %q", got)
	}

	got = strings.Join(decodeInput("\xff"), "\n")
	if !strings.Contains(got, "Raw") {
		t.Fatalf("invalid byte should become Raw: %q", got)
	}
}

func TestToTapeCast(t *testing.T) {
	header := `{"version": 2, "width": 90, "height": 30}
[0.1, "i", "echo one\r"]
[0.4, "o", "one\r\n"]
[0.9, "m", "setup done"]
`
	file := header + "\n[1.0, \"i\", \"\\u001b[C\"]\n[1.1, \"r\", \"100x25\"]\n"
	path := t.TempDir() + "/rec.cast"
	if err := os.WriteFile(path, []byte(file), 0644); err != nil {
		t.Fatal(err)
	}
	cast, err := Parse(path)
	if err != nil {
		t.Fatal(err)
	}
	tape := cast.ToTapeCast()

	for _, want := range []string{
		"Set Width 90", "Set Height 30",
		"Type \"echo one\"", "Enter", "WaitStable @250ms",
		"# marker: setup done", "Right", "Resize 100 25",
	} {
		if !strings.Contains(tape, want) {
			t.Fatalf("tape missing %q:\n%s", want, tape)
		}
	}
}

func TestOscInputIgnored(t *testing.T) {
	got := strings.Join(decodeInput("\x1b]0;title\x07ls"), "\n")
	if got != "Type \"ls\"" {
		t.Fatalf("OSC sequences should be dropped: %q", got)
	}
}
