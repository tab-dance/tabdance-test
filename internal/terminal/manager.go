package terminal

import (
	"fmt"
	"sync"
)

type TerminalManager struct {
	mu        sync.Mutex
	terminals map[string]*ManagedTerminal
	counter   int
}

func NewTerminalManager() *TerminalManager {
	return &TerminalManager{
		terminals: make(map[string]*ManagedTerminal),
	}
}

func (m *TerminalManager) Create(cfg CreateConfig) (*ManagedTerminal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counter++
	id := fmt.Sprintf("term_%d", m.counter)

	mt, err := NewManagedTerminal(id, cfg)
	if err != nil {
		return nil, err
	}

	m.terminals[id] = mt
	return mt, nil
}

func (m *TerminalManager) Get(id string) (*ManagedTerminal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	mt, ok := m.terminals[id]
	if !ok {
		return nil, fmt.Errorf("terminal %q not found", id)
	}
	return mt, nil
}

type TerminalInfo struct {
	ID      string `json:"id"`
	Cols    uint16 `json:"cols"`
	Rows    uint16 `json:"rows"`
	Command string `json:"command"`
}

func (m *TerminalManager) List() []TerminalInfo {
	m.mu.Lock()
	defer m.mu.Unlock()

	infos := make([]TerminalInfo, 0, len(m.terminals))
	for _, mt := range m.terminals {
		mt.mu.Lock()
		infos = append(infos, TerminalInfo{
			ID:      mt.ID,
			Cols:    mt.Cols,
			Rows:    mt.Rows,
			Command: mt.Command,
		})
		mt.mu.Unlock()
	}
	return infos
}

func (m *TerminalManager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	mt, ok := m.terminals[id]
	if !ok {
		return fmt.Errorf("terminal %q not found", id)
	}

	mt.Close()
	delete(m.terminals, id)
	return nil
}

func (m *TerminalManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, mt := range m.terminals {
		mt.Close()
		delete(m.terminals, id)
	}
}
