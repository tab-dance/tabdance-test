package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	libghostty "go.mitchellh.com/libghostty"

	"github.com/creack/pty"
	"github.com/tab-dance/tabdance-test/internal/recording"
)

type ManagedTerminal struct {
	ID           string
	Terminal     *libghostty.Terminal
	KeyEncoder   *libghostty.KeyEncoder
	MouseEncoder *libghostty.MouseEncoder
	Cmd          *exec.Cmd
	Ptmx         *os.File
	recPtr       atomic.Pointer[recording.AsciicastRecorder]
	Command      string
	Cols         uint16
	Rows         uint16

	mu          sync.Mutex
	ptmxPtr     *atomic.Pointer[os.File]
	done        chan struct{}
	close       sync.Once
	subscribers []*outputSubscriber
	// lastOutputTime is when the child last wrote output; used by
	// quiet-period waits.
	lastOutputTime time.Time
	// history is the last maxOutputHistory bytes written by the child.
	history []byte
	// osc tracks semantic prompt markers (OSC 133) seen on the stream.
	osc        osc133State
	oscScanner osc133Scanner
	// Child exit state, recorded once cmd.Wait returns in startPTY.
	exited     bool
	exitCode   int
	exitSignal string
}

// maxOutputHistory caps the ring buffer of recent child output kept per
// terminal for failure diagnostics and the output command.
const maxOutputHistory = 32 * 1024

type CreateConfig struct {
	Cols             uint16
	Rows             uint16
	Scrollback       uint
	Command          string
	Env              map[string]string
	WorkingDirectory string
}

func NewManagedTerminal(id string, cfg CreateConfig) (*ManagedTerminal, error) {
	// The terminal's responses to shell queries (cursor position, device
	// attributes, ...) are written to the PTY synchronously during VTWrite:
	// the query is processed in the read loop and the response reaches the
	// shell before the next output chunk is read. Buffering them through a
	// separate write loop makes the response arrival time race with the
	// shell's rendering state machine, producing non-deterministic output.
	var ptmx atomic.Pointer[os.File]
	term, err := libghostty.NewTerminal(
		libghostty.WithSize(cfg.Cols, cfg.Rows),
		libghostty.WithMaxScrollbackLines(cfg.Scrollback),
		libghostty.WithWritePty(func(_ *libghostty.Terminal, data []byte) {
			f := ptmx.Load()
			if f != nil {
				f.Write(data)
			}
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("create terminal: %w", err)
	}

	keyEnc, err := libghostty.NewKeyEncoder()
	if err != nil {
		term.Close()
		return nil, fmt.Errorf("create key encoder: %w", err)
	}
	keyEnc.SetFromTerminal(term)

	mouseEnc, err := libghostty.NewMouseEncoder()
	if err != nil {
		keyEnc.Close()
		term.Close()
		return nil, fmt.Errorf("create mouse encoder: %w", err)
	}
	mouseEnc.SetFromTerminal(term)

	m := &ManagedTerminal{
		ID:           id,
		Terminal:     term,
		KeyEncoder:   keyEnc,
		MouseEncoder: mouseEnc,
		Command:      cfg.Command,
		Cols:         cfg.Cols,
		Rows:         cfg.Rows,
		done:         make(chan struct{}),
	}

	m.ptmxPtr = &ptmx

	if cfg.Command != "" {
		if err := m.startPTY(cfg); err != nil {
			m.Close()
			return nil, fmt.Errorf("start pty: %w", err)
		}
	}

	return m, nil
}

func (m *ManagedTerminal) startPTY(cfg CreateConfig) error {
	cmd := exec.Command("sh", "-c", cfg.Command)
	cmd.Env = os.Environ()
	for k, v := range cfg.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	if cfg.WorkingDirectory != "" {
		cmd.Dir = cfg.WorkingDirectory
	}

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{
		Cols: uint16(cfg.Cols),
		Rows: uint16(cfg.Rows),
	})
	if err != nil {
		return fmt.Errorf("start command: %w", err)
	}

	m.Cmd = cmd
	m.Ptmx = ptmx
	m.ptmxPtr.Store(ptmx)

	go m.ptyReadLoop()
	// Reap the child and record its exit state (code, signal) so tests can
	// assert on it; a zombie never lingers in the daemon.
	go func() {
		_ = cmd.Wait()
		m.recordExit(cmd)
	}()

	return nil
}

func (m *ManagedTerminal) ptyReadLoop() {
	buf := make([]byte, 4096)
	for {
		select {
		case <-m.done:
			return
		default:
		}

		n, err := m.Ptmx.Read(buf)
		if err != nil {
			return
		}
		if n == 0 {
			continue
		}

		data := buf[:n]

		mt := m.Terminal
		rec := m.recPtr.Load()

		m.mu.Lock()
		m.lastOutputTime = time.Now()
		m.history = append(m.history, data...)
		if overflow := len(m.history) - maxOutputHistory; overflow > 0 {
			m.history = m.history[overflow:]
		}
		m.applyOSC133Locked(m.oscScanner.feed(data))
		mt.VTWrite(data)
		if rec != nil {
			rec.WriteOutput(data)
		}
		subs := append([]*outputSubscriber(nil), m.subscribers...)
		m.mu.Unlock()

		// Deliver without holding the mutex so a slow consumer cannot stall
		// VTWrite; buf is reused by the next Read, so the chunk must be
		// copied because queued subscribers may outlive this iteration.
		if len(subs) > 0 {
			chunk := make([]byte, len(data))
			copy(chunk, data)
			for _, sub := range subs {
				sub.deliver(chunk)
			}
		}
	}
}

func (m *ManagedTerminal) Close() {
	m.close.Do(func() {
		close(m.done)

		// Wake any attach consumers blocked in deliver and end their range
		// loops; otherwise a closed terminal leaves them hanging forever.
		m.mu.Lock()
		subs := m.subscribers
		m.subscribers = nil
		m.mu.Unlock()
		for _, sub := range subs {
			sub.closeUnsubscribed()
		}

		if rec := m.recPtr.Load(); rec != nil {
			rec.Close()
		}

		m.KeyEncoder.Close()
		m.MouseEncoder.Close()

		if pgid := m.childPgid(); pgid > 0 {
			// pty.StartWithSize makes the child a session leader (Setsid), so
			// the negative PID targets the whole group including descendants
			// our direct child spawned (`sh -c` expansions, TUI children).
			syscall.Kill(-pgid, syscall.SIGTERM)
			time.Sleep(100 * time.Millisecond)
			syscall.Kill(-pgid, syscall.SIGKILL)
		}

		if m.Ptmx != nil {
			m.ptmxPtr.Store(nil)
			m.Ptmx.Close()
		}

		m.Terminal.Close()
	})
}

func (m *ManagedTerminal) Resize(cols, rows uint16) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.Terminal.Resize(cols, rows, 8, 16); err != nil {
		return err
	}

	m.Cols = cols
	m.Rows = rows

	m.KeyEncoder.SetFromTerminal(m.Terminal)
	m.MouseEncoder.SetFromTerminal(m.Terminal)

	if rec := m.recPtr.Load(); rec != nil {
		rec.WriteResize(int(cols), int(rows))
	}

	if m.Ptmx != nil {
		pty.Setsize(m.Ptmx, &pty.Winsize{
			Cols: uint16(cols),
			Rows: uint16(rows),
		})
	}

	return nil
}

func (m *ManagedTerminal) StartRecording(path string, idleLimit float64) error {
	rec, err := recording.NewAsciicastRecorder(path, int(m.Cols), int(m.Rows), idleLimit)
	if err != nil {
		return err
	}
	if prev := m.recPtr.Swap(rec); prev != nil {
		// Never leak the previous recorder's file handle; swap was atomic so
		// output never misses a writer.
		prev.Close()
	}
	return nil
}

func (m *ManagedTerminal) StopRecording() error {
	rec := m.recPtr.Swap(nil)
	if rec == nil {
		return nil
	}
	return rec.Close()
}

// IsRecording reports whether an asciicast recording is running.
func (m *ManagedTerminal) IsRecording() bool {
	return m.recPtr.Load() != nil
}

// PauseRecording toggles capture on the active recording, matching
// asciinema's mute: while paused, events are dropped and the timeline
// freezes. No-op when nothing is recording.
func (m *ManagedTerminal) PauseRecording(paused bool) {
	if rec := m.recPtr.Load(); rec != nil {
		rec.PauseRecording(paused)
	}
}

func (m *ManagedTerminal) AddMarker(label string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	rec := m.recPtr.Load()
	if rec == nil {
		return fmt.Errorf("not recording")
	}
	return rec.WriteMarker(label)
}

func (m *ManagedTerminal) WritePty(data []byte) (int, error) {
	if m.Ptmx == nil {
		return 0, fmt.Errorf("no PTY available")
	}
	if rec := m.recPtr.Load(); rec != nil {
		rec.WriteInput(data)
	}
	return m.Ptmx.Write(data)
}

func (m *ManagedTerminal) Write(data []byte) (int, error) {
	return m.WritePty(data)
}

func (m *ManagedTerminal) Lock()   { m.mu.Lock() }
func (m *ManagedTerminal) Unlock() { m.mu.Unlock() }

func (m *ManagedTerminal) GetTerminal() *libghostty.Terminal     { return m.Terminal }
func (m *ManagedTerminal) GetKeyEncoder() *libghostty.KeyEncoder { return m.KeyEncoder }
func (m *ManagedTerminal) GetMouseEncoder() *libghostty.MouseEncoder {
	return m.MouseEncoder
}

// subscriberQueue is the delivery buffer per output subscriber.
const subscriberQueue = 256

// subscribeLocked registers a new subscriber; the caller must hold m.mu.
// Registration under the lock is what makes attach streams gapless: every
// output chunk is broadcast either before or after registration, never
// between the two.
func (m *ManagedTerminal) subscribeLocked() *outputSubscriber {
	sub := newOutputSubscriber(subscriberQueue)
	m.subscribers = append(m.subscribers, sub)
	return sub
}

func (m *ManagedTerminal) Subscribe() *outputSubscriber {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.subscribeLocked()
}

func (m *ManagedTerminal) Unsubscribe(sub *outputSubscriber) {
	m.mu.Lock()
	for i, s := range m.subscribers {
		if s == sub {
			m.subscribers = append(m.subscribers[:i], m.subscribers[i+1:]...)
			break
		}
	}
	m.mu.Unlock()
	sub.closeUnsubscribed()
}

// childPgid returns the child's process group ID (its PID, since
// pty.StartWithSize makes it a session leader), or 0 if there is no child.
func (m *ManagedTerminal) childPgid() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Cmd == nil || m.Cmd.Process == nil {
		return 0
	}
	return m.Cmd.Process.Pid
}

// recordExit captures the child's exit code (or termination signal) under
// the terminal mutex. Run from the reaper goroutine after cmd.Wait returns.
func (m *ManagedTerminal) recordExit(cmd *exec.Cmd) {
	ws := cmd.ProcessState
	if ws == nil {
		return
	}

	status, ok := ws.Sys().(syscall.WaitStatus)
	if !ok {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.exited = true
	if status.Signaled() {
		m.exitCode = -1
		m.exitSignal = status.Signal().String()
	} else {
		m.exitCode = status.ExitStatus()
	}
}

// Exited reports whether the child process has terminated.
func (m *ManagedTerminal) Exited() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.exited
}

// ExitState returns the child's termination status: exit code (with -1
// when killed by a signal) and the signal name if killed. The child is
// the `sh -c` wrapper started by startPTY; if the command string itself
// is a shell snippet, the wrapper reports an inner signal death as a
// plain exit code 128+N instead of a signal.
func (m *ManagedTerminal) ExitState() (exited bool, exitCode int, signal string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.exited, m.exitCode, m.exitSignal
}

// RecentOutput returns up to n bytes of the child's most recent output,
// from a ring buffer refreshed by the PTY read loop.
func (m *ManagedTerminal) RecentOutput(n int) []byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	if n <= 0 || n > len(m.history) {
		n = len(m.history)
	}
	out := make([]byte, n)
	copy(out, m.history[len(m.history)-n:])
	return out
}

// LastOutputTime returns when the child last wrote output.
func (m *ManagedTerminal) LastOutputTime() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastOutputTime
}

// applyOSC133Locked updates the terminal's semantic prompt state.
// The caller must hold the terminal lock.
func (m *ManagedTerminal) applyOSC133Locked(ev osc133Events) {
	if ev.Prompt {
		m.osc.prompt = time.Now()
		m.osc.promptCount++
	}
	if ev.CommandEnd {
		m.osc.commandEnd = time.Now()
		m.osc.cmdEndCount++
		if ev.HadCommandExit {
			m.osc.commandExit = ev.CommandExit
			m.osc.hadCommandExit = true
		}
	}
}

// PromptCount returns the number of OSC 133 prompt-start (A) markers seen.
func (m *ManagedTerminal) PromptCount() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.osc.promptCount
}

// CommandEndCount returns the number of OSC 133 command-finished (D)
// markers seen.
func (m *ManagedTerminal) CommandEndCount() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.osc.cmdEndCount
}

// LastPromptTime returns when the last OSC 133 prompt-start marker (A)
// was seen, or the zero time if none.
func (m *ManagedTerminal) LastPromptTime() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.osc.prompt
}

// LastCommandEnd returns when the last OSC 133 command-finished marker
// (D) was seen plus its exit code when present.
func (m *ManagedTerminal) LastCommandEnd() (t time.Time, ok bool, exitCode int, hasExit bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.osc.commandEnd, !m.osc.commandEnd.IsZero(), m.osc.commandExit, m.osc.hadCommandExit
}
