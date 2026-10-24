package service

import (
	"fmt"

	"github.com/tab-dance/tabdance-test/internal/terminal"
)

// CreateResult is the response shape for terminal creation.
type CreateResult struct {
	TerminalID string `json:"id"`
	Cols       uint16 `json:"cols"`
	Rows       uint16 `json:"rows"`
}

// Exec creates a terminal, optionally starting a recording, and tears it
// down again if the recorder cannot be attached.
func (s *Service) Exec(cfg terminal.CreateConfig, recordPath string, idleTimeLimit float64) (*CreateResult, error) {
	mt, err := s.terminals.Create(cfg)
	if err != nil {
		return nil, err
	}

	if recordPath != "" {
		if err := mt.StartRecording(recordPath, idleTimeLimit); err != nil {
			s.terminals.Delete(mt.ID)
			return nil, fmt.Errorf("start recording: %w", err)
		}
	}

	return &CreateResult{TerminalID: mt.ID, Cols: mt.Cols, Rows: mt.Rows}, nil
}

func (s *Service) List() []terminal.TerminalInfo {
	return s.terminals.List()
}

// InspectResult describes a terminal, including child exit status when
// the child has already terminated.
type InspectResult struct {
	ID         string `json:"id"`
	Cols       uint16 `json:"cols"`
	Rows       uint16 `json:"rows"`
	Command    string `json:"command"`
	Recording  bool   `json:"recording"`
	Exited     bool   `json:"exited,omitempty"`
	ExitCode   *int   `json:"exit_code,omitempty"`
	ExitSignal string `json:"exit_signal,omitempty"`
}

func (s *Service) Inspect(id string) (*InspectResult, error) {
	mt, err := s.terminals.Get(id)
	if err != nil {
		return nil, err
	}

	mt.Lock()
	cols, rows, command := mt.Cols, mt.Rows, mt.Command
	recording := mt.IsRecording()
	exited, code, signal := mt.ExitState()
	mt.Unlock()

	result := &InspectResult{
		ID:        mt.ID,
		Cols:      cols,
		Rows:      rows,
		Command:   command,
		Recording: recording,
	}
	if exited {
		result.Exited = true
		c := code
		result.ExitCode = &c
		if signal != "" {
			result.ExitSignal = signal
		}
	}
	return result, nil
}

func (s *Service) Delete(id string) error {
	return s.terminals.Delete(id)
}
