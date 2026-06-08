package client

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/tab-dance/tabdance-test/internal/daemon"
)

func EnsureDaemon() error {
	if IsDaemonRunning() {
		return nil
	}

	return startDaemon()
}

func IsDaemonRunning() bool {
	pidPath := daemon.PIDPath()
	pidData, err := os.ReadFile(pidPath)
	if err != nil {
		return false
	}

	pid, err := strconv.Atoi(string(pidData))
	if err != nil {
		return false
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	if err := proc.Signal(syscall.Signal(0)); err != nil {
		os.Remove(pidPath)
		os.Remove(daemon.SocketPath())
		return false
	}

	conn, err := daemonDial()
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func startDaemon() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable: %w", err)
	}

	dir := filepath.Dir(daemon.SocketPath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create socket dir: %w", err)
	}

	cmd := exec.Command(exe, "--daemon")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start daemon: %w", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
		if IsDaemonRunning() {
			return nil
		}
	}

	return fmt.Errorf("daemon did not start within timeout")
}

func daemonDial() (net.Conn, error) {
	return net.DialTimeout("unix", daemon.SocketPath(), 2*time.Second)
}
