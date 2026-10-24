package recording

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type AsciicastRecorder struct {
	file      *os.File
	enc       *json.Encoder
	start     time.Time
	idleLimit float64
	lastWrite float64
	// mu serializes event writes: output events come from the PTY read
	// loop, input events from WritePty callers in other goroutines.
	mu sync.Mutex
}

func NewAsciicastRecorder(path string, width, height int, idleLimit float64) (*AsciicastRecorder, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	r := &AsciicastRecorder{
		file:      f,
		enc:       json.NewEncoder(f),
		start:     time.Now(),
		idleLimit: idleLimit,
	}

	header := map[string]any{
		"version":         2,
		"width":           width,
		"height":          height,
		"env":             map[string]string{"TERM": "xterm-ghostty"},
		"idle_time_limit": idleLimit,
	}
	if err := r.enc.Encode(header); err != nil {
		f.Close()
		return nil, err
	}

	return r, nil
}

func (r *AsciicastRecorder) elapsedLocked() float64 {
	raw := time.Since(r.start).Seconds()
	if r.idleLimit > 0 && r.lastWrite > 0 {
		gap := raw - r.lastWrite
		if gap > r.idleLimit {
			raw = r.lastWrite + r.idleLimit
		}
	}
	r.lastWrite = raw
	return raw
}

func (r *AsciicastRecorder) WriteOutput(data []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.enc.Encode([]any{r.elapsedLocked(), "o", string(data)})
}

func (r *AsciicastRecorder) WriteInput(data []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.enc.Encode([]any{r.elapsedLocked(), "i", string(data)})
}

func (r *AsciicastRecorder) WriteMarker(label string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.enc.Encode([]any{r.elapsedLocked(), "m", label})
}

func (r *AsciicastRecorder) WriteResize(cols, rows int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.enc.Encode([]any{r.elapsedLocked(), "r", fmt.Sprintf("%dx%d", cols, rows)})
}

func (r *AsciicastRecorder) Close() error {
	return r.file.Close()
}
