package daemon

import (
	"fmt"
	"os"
	"syscall"
)

// acquireInstanceLock takes an exclusive, non-blocking flock on lockPath.
// It returns the open file (kept locked for the daemon's lifetime, released
// on close) or an error if another daemon is already running. This prevents
// two competing daemon processes from stealing each other's socket file.
func acquireInstanceLock(lockPath string) (*os.File, error) {
	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("open lock file: %w", err)
	}

	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		lock.Close()
		return nil, fmt.Errorf("another daemon is already running")
	}

	return lock, nil
}
