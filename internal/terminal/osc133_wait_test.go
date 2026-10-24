package terminal

import (
	"strings"
	"testing"
	"time"

	"github.com/tab-dance/tabdance-test/internal/tape"
)

func oscFeedMarker(t *testing.T, mt *ManagedTerminal, seq string) {
	t.Helper()
	mt.mu.Lock()
	fe := mt.oscScanner.feed([]byte(seq))
	mt.applyOSC133Locked(fe)
	mt.mu.Unlock()
}

func TestWaitPromptRequiresNewMarker(t *testing.T) {
	mt, err := NewManagedTerminal("prompt-wait", CreateConfig{Cols: 20, Rows: 5, Command: "sleep 300"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	t.Cleanup(mt.Close)

	oscFeedMarker(t, mt, "\x1b]133;A\x07")

	if err := tape.NewTapeExecutor(mt, 5*time.Second).ExecuteString("WaitPrompt @150ms\n"); err == nil ||
		!strings.Contains(err.Error(), "no OSC 133 prompt marker") {
		t.Fatalf("expected timeout on stale marker, got: %v", err)
	}

	go func() {
		time.Sleep(30 * time.Millisecond)
		oscFeedMarker(t, mt, "\x1b]133;A\x07")
	}()
	if err := tape.NewTapeExecutor(mt, 5*time.Second).ExecuteString("WaitPrompt @2s\n"); err != nil {
		t.Fatalf("WaitPrompt should pass on fresh marker: %v", err)
	}
}

func TestWaitCommandExitCode(t *testing.T) {
	mt, err := NewManagedTerminal("cmd-wait", CreateConfig{Cols: 20, Rows: 5, Command: "sleep 300"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	t.Cleanup(mt.Close)

	oscFeedMarker(t, mt, "\x1b]133;D;7\x07")
	err = tape.NewTapeExecutor(mt, 5*time.Second).ExecuteString("WaitCommand 0 @150ms\n")
	if err == nil || !strings.Contains(err.Error(), "no OSC 133 command-finished marker with status 0") {
		t.Fatalf("expected stale-marker timeout, got: %v", err)
	}

	// Fresh marker with a different code is a hard error.
	go func() {
		time.Sleep(30 * time.Millisecond)
		oscFeedMarker(t, mt, "\x1b]133;D;7\x07")
	}()
	err = tape.NewTapeExecutor(mt, 5*time.Second).ExecuteString("WaitCommand 0 @2s\n")
	if err == nil || !strings.Contains(err.Error(), "exited with status 7, expected 0") {
		t.Fatalf("expected code mismatch error, got: %v", err)
	}

	// Fresh marker with the expected code passes.
	mt2, err2 := NewManagedTerminal("cmd-wait2", CreateConfig{Cols: 20, Rows: 5, Command: "sleep 300"})
	if err2 != nil {
		t.Fatalf("create: %v", err2)
	}
	t.Cleanup(mt2.Close)
	go func() {
		time.Sleep(30 * time.Millisecond)
		oscFeedMarker(t, mt2, "\x1b]133;D;7\x07")
	}()
	if err := tape.NewTapeExecutor(mt2, 5*time.Second).ExecuteString("WaitCommand 7 @2s\n"); err != nil {
		t.Fatalf("matching exit code should pass, got: %v", err)
	}

	// The no-code form passes on any fresh marker.
	mt3, err3 := NewManagedTerminal("cmd-wait3", CreateConfig{Cols: 20, Rows: 5, Command: "sleep 300"})
	if err3 != nil {
		t.Fatalf("create: %v", err3)
	}
	t.Cleanup(mt3.Close)
	go func() {
		time.Sleep(30 * time.Millisecond)
		oscFeedMarker(t, mt3, "\x1b]133;D;1\x07")
	}()
	if err := tape.NewTapeExecutor(mt3, 5*time.Second).ExecuteString("WaitCommand @2s\n"); err != nil {
		t.Fatalf("no-code WaitCommand should pass: %v", err)
	}

}
