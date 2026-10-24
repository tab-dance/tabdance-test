package terminal

import (
	"bytes"
	"syscall"
	"testing"
	"time"
)

func newTestTerminal(t *testing.T, id, command string) *ManagedTerminal {
	t.Helper()
	mt, err := NewManagedTerminal(id, CreateConfig{Cols: 20, Rows: 5, Command: command})
	if err != nil {
		t.Fatalf("create terminal: %v", err)
	}
	t.Cleanup(mt.Close)
	return mt
}

func waitForExit(t *testing.T, mt *ManagedTerminal, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		if exited, _, _ := mt.ExitState(); exited {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("child never exited")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestExitStateCaptured(t *testing.T) {
	mt := newTestTerminal(t, "exit3", "sh -c 'exit 3'")
	waitForExit(t, mt, 5*time.Second)

	exited, code, signal := mt.ExitState()
	if !exited || code != 3 || signal != "" {
		t.Fatalf("expected exited code=3, got exited=%v code=%d signal=%q", exited, code, signal)
	}
}

func TestExitStateCapturesSignal(t *testing.T) {
	mt := newTestTerminal(t, "killer", "sh -c 'kill -9 $$'")
	waitForExit(t, mt, 5*time.Second)

	exited, code, signal := mt.ExitState()
	if !exited || code != -1 || signal != "killed" {
		t.Fatalf("expected signal killed, got exited=%v code=%d signal=%q", exited, code, signal)
	}
}

func TestRecentOutputHistory(t *testing.T) {
	mt := newTestTerminal(t, "hist", "sh -c 'echo history-test-token'; exit 0")
	waitForExit(t, mt, 5*time.Second)

	recent := mt.RecentOutput(4096)
	if len(recent) == 0 {
		t.Fatal("expected recent output, got none")
	}
	if !bytesContains(recent, "history-test-token") {
		t.Fatalf("history missing token: %q", string(recent))
	}
}

func bytesContains(b []byte, sub string) bool {
	return bytes.Contains(b, []byte(sub))
}

func TestRecentOutputWindow(t *testing.T) {
	mt := newTestTerminal(t, "win", "sh -c 'echo AAAAAAAA; echo BBBBBBBB'; exit 0")
	waitForExit(t, mt, 5*time.Second)

	recent := mt.RecentOutput(10)
	if len(recent) > 10 {
		t.Fatalf("History window exceeded: %d bytes", len(recent))
	}
}

func TestCloseKillsProcessGroup(t *testing.T) {
	mt, err := NewManagedTerminal("group", CreateConfig{
		Cols: 20, Rows: 5, Command: "sh -c 'sleep 300'",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	pid := mt.Cmd.Process.Pid
	if err := syscall.Kill(pid, syscall.Signal(0)); err != nil {
		t.Fatalf("child not running before close: %v", err)
	}

	mt.Close()

	// Kill escalates: SIGTERM escalation, then SIGKILL, should take the
	// child out within the internal close grace.
	deadline := time.Now().Add(3 * time.Second)
	for {
		err := syscall.Kill(pid, syscall.Signal(0))
		if err == syscall.ESRCH {
			return // gone
		}
		if time.Now().After(deadline) {
			t.Fatalf("child survived close: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func TestLastOutputTimeTracking(t *testing.T) {
	mt := newTestTerminal(t, "lastout", "sh -c 'echo tick'; exit 0")
	waitForExit(t, mt, 5*time.Second)

	if mt.LastOutputTime().IsZero() {
		t.Fatal("expected lastOutputTime to be set after child output")
	}
}
