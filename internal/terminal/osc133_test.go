package terminal

import (
	"testing"
)

func TestOSC133ScannerWholeBuffer(t *testing.T) {
	var s osc133Scanner
	ev := s.feed([]byte("\x1b]133;A\x07\x1b]133;D;7\x1b\\"))
	if !ev.Prompt || !ev.CommandEnd || !ev.HadCommandExit || ev.CommandExit != 7 {
		t.Fatalf("unexpected events: %+v", ev)
	}
}

func TestOSC133ScannerSplitSequences(t *testing.T) {
	var s osc133Scanner
	feeds := []osc133Events{
		s.feed([]byte("output...\x1b]133")),
		s.feed([]byte(";A\x07next\x1b]133;")),
		s.feed([]byte("D42")),
		s.feed([]byte("\x1b\\")),
	}

	ents := append([]osc133Events{}, feeds...)
	if ents[0].Prompt {
		t.Fatal("prompt must not fire on a split marker")
	}
	if !ents[1].Prompt {
		t.Fatal("prompt should fire when the BEL completes the split marker")
	}
	if !ents[3].CommandEnd {
		t.Fatal("command end should fire when the sequence completes")
	}

}

func TestOSC133ScannerIntegratesWithTerminal(t *testing.T) {
	mt, err := NewManagedTerminal("osc", CreateConfig{Cols: 20, Rows: 5})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	t.Cleanup(mt.Close)

	// Inject output through the same path the read loop uses.
	mt.mu.Lock()
	fe := mt.oscScanner.feed([]byte("\x1b]133;A\x07someshell\x1b]133;A\x07"))
	mt.applyOSC133Locked(fe)
	fe = mt.oscScanner.feed([]byte("users \x1b]133;D;2\x07prompt"))
	mt.applyOSC133Locked(fe)
	mt.mu.Unlock()

	if mt.LastPromptTime().IsZero() {
		t.Fatal("expected prompt time set")
	}
	if _, ok, code, hasExit := mt.LastCommandEnd(); !ok || code != 2 || !hasExit {
		t.Fatalf("unexpected command end: ok=%v code=%d hasExit=%v", ok, code, hasExit)
	}
}
