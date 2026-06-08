package daemon

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/tab-dance/tabdance-test/internal/terminal"
)

const (
	defaultSocketDir   = "tabdance-test"
	defaultSocketName  = "daemon.sock"
	defaultPIDFileName = "daemon.pid"
	defaultIdleTimeout = 30 * time.Second
)

type Daemon struct {
	manager     *terminal.TerminalManager
	socketPath  string
	pidPath     string
	idleTimeout time.Duration

	mu          sync.Mutex
	listener    net.Listener
	done        chan struct{}
	clients     int
	lastActive  time.Time
}

func socketDir() string {
	if dir := os.Getenv("TABDANCE_TEST_SOCK"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", defaultSocketDir)
}

func SocketPath() string {
	return filepath.Join(socketDir(), defaultSocketName)
}

func PIDPath() string {
	return filepath.Join(socketDir(), defaultPIDFileName)
}

func New() *Daemon {
	dir := socketDir()
	return &Daemon{
		manager:     terminal.NewTerminalManager(),
		socketPath:   filepath.Join(dir, defaultSocketName),
		pidPath:     filepath.Join(dir, defaultPIDFileName),
		idleTimeout: idleTimeout(),
		done:        make(chan struct{}),
	}
}

func idleTimeout() time.Duration {
	if s := os.Getenv("TABDANCE_TEST_IDLE_TIMEOUT"); s != "" {
		if d, err := time.ParseDuration(s); err == nil {
			return d
		}
	}
	return defaultIdleTimeout
}

func (d *Daemon) Run() error {
	if err := os.MkdirAll(filepath.Dir(d.socketPath), 0755); err != nil {
		return fmt.Errorf("create socket dir: %w", err)
	}

	os.Remove(d.socketPath)

	l, err := net.Listen("unix", d.socketPath)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	d.listener = l

	if err := os.WriteFile(d.pidPath, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		l.Close()
		os.Remove(d.socketPath)
		return fmt.Errorf("write pid: %w", err)
	}

	defer func() {
		os.Remove(d.pidPath)
		os.Remove(d.socketPath)
		d.manager.CloseAll()
	}()

	go d.idleWatch()

	for {
		conn, err := l.Accept()
		if err != nil {
			select {
			case <-d.done:
				return nil
			default:
				return fmt.Errorf("accept: %w", err)
			}
		}

		d.mu.Lock()
		d.clients++
		d.mu.Unlock()

		go d.handleConn(conn)
	}
}

func (d *Daemon) idleWatch() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-d.done:
			return
		case <-ticker.C:
			d.mu.Lock()
			count := len(d.manager.List())
			clients := d.clients
			d.mu.Unlock()

			if count == 0 && clients == 0 {
				if d.lastActive.IsZero() {
					d.lastActive = time.Now()
				} else if time.Since(d.lastActive) > d.idleTimeout {
					d.listener.Close()
					close(d.done)
					return
				}
			} else {
				d.lastActive = time.Time{}
			}
		}
	}
}

func (d *Daemon) handleConn(conn net.Conn) {
	defer func() {
		conn.Close()
		d.mu.Lock()
		d.clients--
		d.mu.Unlock()
	}()

	dec := json.NewDecoder(conn)
	enc := json.NewEncoder(conn)

	for {
		var req Request
		if err := dec.Decode(&req); err != nil {
			return
		}

		if req.Method == "attach" {
			d.handleAttach(req, conn)
			return
		}

		resp := d.handleRequest(req)

		if err := enc.Encode(resp); err != nil {
			return
		}
	}
}

func (d *Daemon) handleRequest(req Request) Response {
	switch req.Method {
	case "exec":
		return d.handleExec(req)
	case "ls":
		return d.handleLs(req)
	case "inspect":
		return d.handleInspect(req)
	case "kill":
		return d.handleKill(req)
	case "rm":
		return d.handleRm(req)
	case "view":
		return d.handleView(req)
	case "cursor":
		return d.handleCursor(req)
	case "size":
		return d.handleSize(req)
	case "cell":
		return d.handleCell(req)
	case "send":
		return d.handleSend(req)
	case "keys":
		return d.handleKeys(req)
	case "mouse":
		return d.handleMouse(req)
	case "paste":
		return d.handlePaste(req)
	case "focus":
		return d.handleFocus(req)
	case "resize":
		return d.handleResize(req)
	case "tape":
		return d.handleTape(req)
	case "wait":
		return d.handleWait(req)
	case "verify":
		return d.handleVerify(req)
	case "snapshot":
		return d.handleSnapshot(req)
	case "record":
		return d.handleRecord(req)
	case "stop":
		return d.handleStop(req)
	case "marker":
		return d.handleMarker(req)
	case "ping":
		return SuccessResponse("pong")
	default:
		return ErrorResponse(fmt.Sprintf("unknown method: %s", req.Method))
	}
}
