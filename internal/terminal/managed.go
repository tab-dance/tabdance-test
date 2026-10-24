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
	Recorder     *recording.AsciicastRecorder
	Command      string
	Cols         uint16
	Rows         uint16

	mu          sync.Mutex
	ptmxPtr     *atomic.Pointer[os.File]
	done        chan struct{}
	close       sync.Once
	subscribers []*outputSubscriber
}

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
	keyEnc.SetOptFromTerminal(term)

	mouseEnc, err := libghostty.NewMouseEncoder()
	if err != nil {
		keyEnc.Close()
		term.Close()
		return nil, fmt.Errorf("create mouse encoder: %w", err)
	}
	mouseEnc.SetOptFromTerminal(term)

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
	// Reap the child process once it exits (or is killed on Close) so it
	// does not linger as a zombie in the daemon.
	go func() {
		_ = cmd.Wait()
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

		m.mu.Lock()
		m.Terminal.VTWrite(data)
		if m.Recorder != nil {
			m.Recorder.WriteOutput(data)
		}
		for _, sub := range m.subscribers {
			select {
			case sub.ch <- data:
			default:
			}
		}
		m.mu.Unlock()
	}
}

func (m *ManagedTerminal) Close() {
	m.close.Do(func() {
		close(m.done)

		if m.Recorder != nil {
			m.Recorder.Close()
		}

		m.KeyEncoder.Close()
		m.MouseEncoder.Close()

		if m.Cmd != nil && m.Cmd.Process != nil {
			m.Cmd.Process.Signal(syscall.SIGTERM)
			time.Sleep(100 * time.Millisecond)
			m.Cmd.Process.Kill()
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

	m.KeyEncoder.SetOptFromTerminal(m.Terminal)
	m.MouseEncoder.SetOptFromTerminal(m.Terminal)

	if m.Recorder != nil {
		m.Recorder.WriteResize(int(cols), int(rows))
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
	m.mu.Lock()
	defer m.mu.Unlock()

	rec, err := recording.NewAsciicastRecorder(path, int(m.Cols), int(m.Rows), idleLimit)
	if err != nil {
		return err
	}
	m.Recorder = rec
	return nil
}

func (m *ManagedTerminal) StopRecording() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Recorder == nil {
		return nil
	}
	err := m.Recorder.Close()
	m.Recorder = nil
	return err
}

func (m *ManagedTerminal) AddMarker(label string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Recorder == nil {
		return fmt.Errorf("not recording")
	}
	return m.Recorder.WriteMarker(label)
}

func (m *ManagedTerminal) WritePty(data []byte) (int, error) {
	if m.Ptmx == nil {
		return 0, fmt.Errorf("no PTY available")
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

func (m *ManagedTerminal) Subscribe() *outputSubscriber {
	sub := &outputSubscriber{ch: make(chan []byte, 64)}
	m.mu.Lock()
	m.subscribers = append(m.subscribers, sub)
	m.mu.Unlock()
	return sub
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
	close(sub.ch)
}
