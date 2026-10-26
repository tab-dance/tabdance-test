package recording

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCloseWritesTrailingIdleEvent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rec.cast")
	r, err := NewAsciicastRecorder(path, 80, 24, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.WriteOutput([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	time.Sleep(120 * time.Millisecond)
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}

	header, events := readCast(t, path)
	if version, _ := header["version"].(float64); version != 3 {
		t.Fatalf("header version = %v, want 3", header["version"])
	}
	term, _ := header["term"].(map[string]any)
	if term == nil {
		t.Fatalf("header missing term object: %v", header)
	}
	if cols, _ := term["cols"].(float64); int(cols) != 80 {
		t.Fatalf("term.cols = %v, want 80", term["cols"])
	}
	if rows, _ := term["rows"].(float64); int(rows) != 24 {
		t.Fatalf("term.rows = %v, want 24", term["rows"])
	}

	var total float64
	for _, ev := range events {
		ts, _ := ev[0].(float64)
		total += ts
	}
	if total < 0.1 {
		t.Fatalf("cumulative intervals must cover the trailing quiet time, got %v", total)
	}
	last := events[len(events)-1]
	if kind, _ := last[1].(string); kind != "o" {
		t.Fatalf("final event type = %v, want o", last[1])
	}
	if data, _ := last[2].(string); data != "" {
		t.Fatalf("final event data = %q, want empty", last[2])
	}
	lone, _ := events[0][2].(string)
	if lone != "hello" {
		t.Fatalf("first output event data = %q, want hello", lone)
	}
}

func TestIdleTimeLimitTrimsIntervals(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rec.cast")
	r, err := NewAsciicastRecorder(path, 80, 24, 0.05)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.WriteOutput([]byte("a")); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	if err := r.WriteOutput([]byte("b")); err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}

	_, events := readCast(t, path)
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	second, _ := events[1][0].(float64)
	if second > 0.06 {
		t.Fatalf("idle-limited interval = %v, want <= 0.05s", second)
	}
	if second < 0.04 {
		t.Fatalf("idle-limited interval = %v, want the full limit", second)
	}
}

func readCast(t *testing.T, path string) (map[string]any, [][]any) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var header map[string]any
	var events [][]any
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		if header == nil {
			if err := json.Unmarshal(line, &header); err != nil {
				t.Fatalf("header line: %v", err)
			}
			continue
		}
		var ev []any
		if err := json.Unmarshal(line, &ev); err != nil {
			t.Fatalf("event line: %v", err)
		}
		if len(ev) != 3 {
			t.Fatalf("unexpected event arity: %v", ev)
		}
		events = append(events, ev)
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if header == nil {
		t.Fatal("missing header")
	}
	return header, events
}
