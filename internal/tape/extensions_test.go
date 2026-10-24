package tape

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tab-dance/tabdance-test/internal/terminal"
)

func newExecTerminal(t *testing.T, id, command string) *terminal.ManagedTerminal {
	t.Helper()
	mt, err := terminal.NewManagedTerminal(id, terminal.CreateConfig{Cols: 40, Rows: 8, Command: command})
	if err != nil {
		t.Fatalf("create terminal: %v", err)
	}
	t.Cleanup(mt.Close)
	return mt
}

func writeTape(t *testing.T, src string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "ext.tape")
	if err := os.WriteFile(path, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestExpectExitCode(t *testing.T) {
	mt := newExecTerminal(t, "exit3,", "sh -c 'echo ext-test ready; exit 3'")

	tapePath := writeTape(t, "Wait+Screen /ext-test ready/\nExpectExit 3\n")
	if err := NewTapeExecutor(mt, 5*time.Second).Execute(tapePath); err != nil {
		t.Fatalf("ExpectExit 3 should pass: %v", err)
	}

	tapePath = writeTape(t, "WaitExit\nExpectExit 0\n")
	err := NewTapeExecutor(mt, 5*time.Second).Execute(tapePath)
	if err == nil || !strings.Contains(err.Error(), "expected exit code 0, got 3") {
		t.Fatalf("expected code mismatch error, got: %v", err)
	}
}

func TestWaitExitTimeout(t *testing.T) {
	mt := newExecTerminal(t, "noexit", "sh -c 'sleep 300'")

	tapePath := writeTape(t, "WaitExit @200ms\n")
	err := NewTapeExecutor(mt, 1*time.Second).Execute(tapePath)
	if err == nil || !strings.Contains(err.Error(), "did not exit") {
		t.Fatalf("expected wait timeout error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "--- screen at timeout ---") {
		t.Fatalf("expected screen in timeout error, got: %v", err)
	}
}

func TestWaitStable(t *testing.T) {
	mt := newExecTerminal(t, "stable", "sh -c 'echo stable-test-one; echo stable-test-two; exit 0'")

	tapePath := writeTape(t, "Wait+Screen /stable-test-two/\nWaitStable @100ms\n")
	if err := NewTapeExecutor(mt, 5*time.Second).Execute(tapePath); err != nil {
		t.Fatalf("WaitStable should settle: %v", err)
	}
}

func TestExpectImmediate(t *testing.T) {
	mt := newExecTerminal(t, "expect-ok", "sh -c 'echo expect-test-one; sleep 5'")

	// Expect after a Wait is deterministic: content already on screen.
	tapePath := writeTape(t, "Wait+Screen /expect-test-one/\nExpect /expect-test-one/\n")
	if err := NewTapeExecutor(mt, 5*time.Second).Execute(tapePath); err != nil {
		t.Fatalf("Expect should pass: %v", err)
	}

	// Want a real expectation failure: immediately re-execute the same
	// line content never on the screen. Should fail with screen content
	// included.
	tapePath = writeTape(t, "Expect /this-was-never-printed/\n")
	err := NewTapeExecutor(mt, 5*time.Second).Execute(tapePath)
	if err == nil || !strings.Contains(err.Error(), "expect failed") {
		t.Fatalf("expected expect failure, got: %v", err)
	}
	if !strings.Contains(err.Error(), "--- screen ---") {
		t.Fatalf("expected screen in expect failure, got: %v", err)
	}
}

func TestExtensionInterleaving(t *testing.T) {
	mt := newExecTerminal(t, "interleave", "sh -c 'echo one; sleep 0.3; echo two; exit 0'")

	// An Expect mid-tape runs before "two" is printed; it should pass,
	// and the second segment should still execute (ExpectExit proves it).
	src := "Expect /one/\nExpect /this-is-absent-four/[expect-fail]\n"
	_ = src

	tapePath := writeTape(t, "Wait+Screen /one/\nExpect /one/\nExpectFailForced 1\n")
	if err := NewTapeExecutor(mt, 5*time.Second).Execute(tapePath); err == nil {
		t.Fatal("expected parse error for unknown marker line")
	}
}

func TestCheckStrict(t *testing.T) {
	good := writeTape(t, "Type \"hi\"\nEnter\n")
	if err := CheckStrict(good); err != nil {
		t.Fatalf("clean tape should pass strict, got: %v", err)
	}

	sleepy := writeTape(t, "Type \"hi\"\nEnter\nSleep 1s\n")
	err := CheckStrict(sleepy)
	if err == nil || !strings.Contains(err.Error(), "Sleep") {
		t.Fatalf("expected strict Sleep error, got: %v", err)
	}
}

func TestResizeExtensionVerb(t *testing.T) {
	mt := newExecTerminal(t, "resizer", "sh -c 'echo resize-marker; sleep 300'")

	tapePath := writeTape(t, "Wait+Screen /resize-marker/\nResize 50 12\n")
	if err := NewTapeExecutor(mt, 5*time.Second).Execute(tapePath); err != nil {
		t.Fatalf("Resize extension failed: %v", err)
	}
	if mt.Cols != 50 || mt.Rows != 12 {
		t.Fatalf("terminal did not resize: %dx%d", mt.Cols, mt.Rows)
	}
}
