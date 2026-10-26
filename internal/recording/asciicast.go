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
	// elapsed is the cumulative interval emitted so far (seconds since
	// start, trimmed by idleLimit). Events carry the interval since the
	// previously emitted event (asciicast v3).
	elapsed float64
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
		"version": 3,
		"term":    map[string]any{"cols": width, "rows": height, "type": "xterm-ghostty"},
	}
	// idle_time_limit means "cap inactivity at this many seconds" for
	// players; 0 would clip every gap to zero, so omit it entirely
	// (absent = unbounded playback).
	if idleLimit > 0 {
		header["idle_time_limit"] = idleLimit
	}
	if err := r.enc.Encode(header); err != nil {
		f.Close()
		return nil, err
	}

	return r, nil
}

// intervalLocked returns the seconds since the previously emitted event,
// capped at the idle limit.
func (r *AsciicastRecorder) intervalLocked() float64 {
	gap := time.Since(r.start).Seconds() - r.elapsed
	if r.idleLimit > 0 && r.elapsed > 0 && gap > r.idleLimit {
		gap = r.idleLimit
	}
	r.elapsed += gap
	return gap
}

func (r *AsciicastRecorder) WriteOutput(data []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.enc.Encode([]any{r.intervalLocked(), "o", string(data)})
}

func (r *AsciicastRecorder) WriteInput(data []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.enc.Encode([]any{r.intervalLocked(), "i", string(data)})
}

func (r *AsciicastRecorder) WriteMarker(label string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.enc.Encode([]any{r.intervalLocked(), "m", label})
}

func (r *AsciicastRecorder) WriteResize(cols, rows int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.enc.Encode([]any{r.intervalLocked(), "r", fmt.Sprintf("%dx%d", cols, rows)})
}

func (r *AsciicastRecorder) Close() error {
	r.mu.Lock()
	_ = r.enc.Encode([]any{r.intervalLocked(), "o", ""})
	r.mu.Unlock()
	return r.file.Close()
}
