package terminal

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sync"
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
	PTYOut       *bytes.Buffer
	Recorder     *recording.AsciicastRecorder
	Command      string
	Cols         uint16
	Rows         uint16

	mu     sync.Mutex
	done   chan struct{}
	ptyOut *bytes.Buffer
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
	ptyOut := &bytes.Buffer{}

	term, err := libghostty.NewTerminal(
		libghostty.WithSize(cfg.Cols, cfg.Rows),
		libghostty.WithMaxScrollback(cfg.Scrollback),
		libghostty.WithWritePty(func(_ *libghostty.Terminal, data []byte) {
			ptyOut.Write(data)
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
		PTYOut:       ptyOut,
		Command:      cfg.Command,
		Cols:         cfg.Cols,
		Rows:         cfg.Rows,
		ptyOut:       ptyOut,
		done:         make(chan struct{}),
	}

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

	go m.ptyReadLoop()
	go m.ptyWriteLoop()

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
		m.mu.Unlock()
	}
}

func (m *ManagedTerminal) ptyWriteLoop() {
	buf := make([]byte, 4096)
	for {
		select {
		case <-m.done:
			return
		default:
		}

		n, err := m.ptyOut.Read(buf)
		if err != nil {
			return
		}
		if n == 0 {
			continue
		}

		if m.Ptmx != nil {
			m.Ptmx.Write(buf[:n])
		}
	}
}

func (m *ManagedTerminal) Close() {
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
		m.Ptmx.Close()
	}

	m.Terminal.Close()
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
