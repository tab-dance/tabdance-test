package recording

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type AsciicastRecorder struct {
	file       *os.File
	enc        *json.Encoder
	start      time.Time
	idleLimit  float64
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

	header := map[string]interface{}{
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

func (r *AsciicastRecorder) WriteOutput(data []byte) error {
	elapsed := time.Since(r.start).Seconds()
	return r.enc.Encode([]interface{}{elapsed, "o", string(data)})
}

func (r *AsciicastRecorder) WriteInput(data []byte) error {
	elapsed := time.Since(r.start).Seconds()
	return r.enc.Encode([]interface{}{elapsed, "i", string(data)})
}

func (r *AsciicastRecorder) WriteMarker(label string) error {
	elapsed := time.Since(r.start).Seconds()
	return r.enc.Encode([]interface{}{elapsed, "m", label})
}

func (r *AsciicastRecorder) WriteResize(cols, rows int) error {
	elapsed := time.Since(r.start).Seconds()
	return r.enc.Encode([]interface{}{elapsed, "r", fmt.Sprintf("%dx%d", cols, rows)})
}

func (r *AsciicastRecorder) Close() error {
	return r.file.Close()
}
