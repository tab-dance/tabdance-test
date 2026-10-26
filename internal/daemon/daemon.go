package daemon

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/tab-dance/tabdance-test/internal/service"
	"github.com/tab-dance/tabdance-test/internal/terminal"
)

const (
	defaultSocketDir   = "tabdance-test"
	defaultSocketName  = "daemon.sock"
	defaultPIDFileName = "daemon.pid"
	defaultIdleTimeout = 30 * time.Second
)

type Daemon struct {
	manager      *terminal.TerminalManager
	svc          *service.Service
	socketPath   string
	pidPath      string
	idleTimeout  time.Duration
	instanceLock *os.File

	mu         sync.Mutex
	listener   net.Listener
	done       chan struct{}
	shutdown   sync.Once
	clients    int
	lastActive time.Time
}

// socketDir returns the global socket directory, one shared daemon for all
// projects. Terminals carry a per-project tag instead (SessionKey of the
// creating shell), so lists and completion stay scoped while every
// directory can still find and address the same background service.
// TABDANCE_TEST_SOCK overrides the directory entirely.
func socketDir() string {
	if dir := os.Getenv("TABDANCE_TEST_SOCK"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", defaultSocketDir)
}

// SessionKey namespaces terminals by the creating project (git repo root
// falling back to the working directory of the calling shell). Used as a
// terminal tag at creation time and as the ls scope; nested project
// directories hash to the same key.
func SessionKey() string {
	start, err := os.Getwd()
	if err != nil {
		return "orphan"
	}
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			dir = start // no repo root found; key by the working directory
			break
		}
		dir = parent
	}
	sum := sha256.Sum256([]byte(dir))
	return hex.EncodeToString(sum[:])[:12]
}

func SocketPath() string {
	return filepath.Join(socketDir(), defaultSocketName)
}

func PIDPath() string {
	return filepath.Join(socketDir(), defaultPIDFileName)
}
func New() *Daemon {
	manager := terminal.NewTerminalManager()
	dir := socketDir()
	return &Daemon{
		manager:     manager,
		svc:         service.NewWithManager(manager),
		socketPath:  filepath.Join(dir, defaultSocketName),
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

// idleExitEnabled reports whether the daemon should exit when idle. A
// non-positive TABDANCE_TEST_IDLE_TIMEOUT (e.g. 0) disables the exit:
// the daemon stays up until shut down by signal, at the cost of keeping
// processes and any PTY children alive until shells remove them.
func (d *Daemon) idleExitEnabled() bool { return d.idleTimeout > 0 }

func (d *Daemon) Run() error {
	dir := filepath.Dir(d.socketPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create socket dir: %w", err)
	}

	lock, err := acquireInstanceLock(d.pidPath + ".lock")
	if err != nil {
		return fmt.Errorf("acquire daemon lock: %w", err)
	}
	d.instanceLock = lock
	defer lock.Close()

	os.Remove(d.socketPath)

	l, err := net.Listen("unix", d.socketPath)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	if err := os.Chmod(d.socketPath, 0600); err != nil {
		l.Close()
		os.Remove(d.socketPath)
		return fmt.Errorf("chmod socket: %w", err)
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

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	acceptErrs := make(chan error, 1)
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				acceptErrs <- err
				return
			}

			d.mu.Lock()
			d.clients++
			d.mu.Unlock()

			go d.handleConn(conn)
		}
	}()

	select {
	case err := <-acceptErrs:
		select {
		case <-d.done:
			return nil
		default:
			return fmt.Errorf("accept: %w", err)
		}
	case <-sigCh:
		d.shutdown.Do(func() { close(d.done) })
		l.Close()
		return nil
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
			if !d.idleExitEnabled() {
				continue
			}
			d.mu.Lock()
			count := len(d.manager.List())
			clients := d.clients
			d.mu.Unlock()

			if count == 0 && clients == 0 {
				if d.lastActive.IsZero() {
					d.lastActive = time.Now()
				} else if time.Since(d.lastActive) > d.idleTimeout {
					d.listener.Close()
					d.shutdown.Do(func() { close(d.done) })
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
	case "output":
		return d.handleOutput(req)
	case "marker":
		return d.handleMarker(req)
	case "ping":
		return SuccessResponse("pong")
	default:
		return ErrorResponse(fmt.Sprintf("unknown method: %s", req.Method))
	}
}
