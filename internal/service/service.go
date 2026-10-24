// Package service holds the transport-neutral terminal operations shared
// by the daemon (Unix socket) and the MCP server (stdio). Handlers do
// request/argument shaping; the service owns terminal access, locking,
// and error context so both transports cannot drift apart.
package service

import (
	"github.com/tab-dance/tabdance-test/internal/terminal"
)

type Service struct {
	terminals *terminal.TerminalManager
}

func New() *Service {
	return NewWithManager(terminal.NewTerminalManager())
}

func NewWithManager(m *terminal.TerminalManager) *Service {
	return &Service{terminals: m}
}

// Get exposes the shared manager (completion, external integrations).
func (s *Service) Get(id string) (*terminal.ManagedTerminal, error) {
	return s.terminals.Get(id)
}

// get returns the terminal or nil; err already describes the failure.
func (s *Service) get(id string) (*terminal.ManagedTerminal, error) {
	return s.terminals.Get(id)
}
