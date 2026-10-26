package castconv

import (
	"fmt"
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

func v3Header(cols, rows int) string {
	return fmt.Sprintf(`{"version": 3, "term": {"cols": %d, "rows": %d, "type": "xterm-ghostty"}}
`, cols, rows)
}

func TestParseBasicEvents(t *testing.T) {
	path := writeCast(t, v3Header(80, 24)+`[0.1, "o", "hello\r\n"]
[0.1, "i", "echo hi\r"]
[0.8, "m", "after"]
[1.0, "r", "100x25"]
[0.0, "x", "0"]
`)
	cast, err := Parse(path)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(cast.Events) != 5 {
		t.Fatalf("expected 5 events, got %d", len(cast.Events))
	}
	if cast.Width != 80 || cast.Height != 24 {
		t.Fatalf("header mismatch: %dx%d", cast.Width, cast.Height)
	}
	// intervals become absolute times: 0.1, 0.2, 1.0, 2.0, 2.0
	want := []float64{0.1, 0.2, 1.0, 2.0, 2.0}
	for i, w := range want {
		if cast.Events[i].Time != w {
			t.Errorf("event %d time = %v, want %v", i, cast.Events[i].Time, w)
		}
	}
	if cast.Events[4].Kind != "x" || cast.Events[4].Data != "0" {
		t.Fatalf("exit event lost: %+v", cast.Events[4])
	}
}

func TestParseRejectsV2Cast(t *testing.T) {
	path := writeCast(t, `{"version": 2, "width": 80, "height": 24}
[0.1, "o", "hello"]
`)
	_, err := Parse(path)
	if err == nil || !strings.Contains(err.Error(), "version 2") {
		t.Fatalf("expected unsupported-version error, got %v", err)
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
	header := v3Header(90, 30)
	file := header + `# recorded by tt
[0.1, "i", "echo one\r"]
[0.3, "o", "one\r\n"]
[0.5, "m", "setup done"]
[0.1, "x", "0"]
[0.1, "i", "\u001b[C"]
[0.1, "r", "100x25"]
`
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

func TestTypingSpeedFromRecordingTiming(t *testing.T) {
	header := v3Header(80, 24)
	events := `# recorded by tt
[0.0, "i", "a"]
[0.1, "i", "b"]
[0.1, "i", "c"]
[0.1, "i", "d"]
[0.1, "i", "e"]
[0.1, "i", "\r"]
`
	path := writeCast(t, header+events)
	cast, err := Parse(path)
	if err != nil {
		t.Fatal(err)
	}
	tape := cast.ToTapeCast()
	if !strings.Contains(tape, "Set TypingSpeed 100ms") {
		t.Fatalf("expected 100ms typing speed from ~100ms gaps:\n%s", tape)
	}

	single := writeCast(t, v3Header(80, 24)+`[0.0, "i", "a"]
`)
	cast, err = Parse(single)
	if err != nil {
		t.Fatal(err)
	}
	tape = cast.ToTapeCast()
	if strings.Contains(tape, "Set TypingSpeed") {
		t.Fatalf("single-keystroke recording must not emit Set TypingSpeed:\n%s", tape)
	}
}
