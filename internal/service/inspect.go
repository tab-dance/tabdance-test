package service

import (
	"fmt"

	"github.com/tab-dance/tabdance-test/internal/inspect"
)

func (s *Service) View(id, format string, trim, unwrap bool) (*inspect.ViewResult, error) {
	if format == "" {
		format = "plain"
	}
	mt, err := s.terminals.Get(id)
	if err != nil {
		return nil, err
	}
	mt.Lock()
	defer mt.Unlock()
	return inspect.GetView(mt.Terminal, format, trim, unwrap)
}

func (s *Service) Cursor(id string) (*inspect.CursorResult, error) {
	mt, err := s.terminals.Get(id)
	if err != nil {
		return nil, err
	}
	mt.Lock()
	defer mt.Unlock()
	return inspect.GetCursor(mt.Terminal), nil
}

func (s *Service) Size(id string) (*inspect.SizeResult, error) {
	mt, err := s.terminals.Get(id)
	if err != nil {
		return nil, err
	}
	mt.Lock()
	defer mt.Unlock()
	return inspect.GetSize(mt.Terminal), nil
}

func (s *Service) Cell(id string, x int, y int) (*inspect.CellResult, error) {
	mt, err := s.terminals.Get(id)
	if err != nil {
		return nil, err
	}
	if x < 0 || y < 0 {
		return nil, fmt.Errorf("cell coordinates must be non-negative, got %d,%d", x, y)
	}
	mt.Lock()
	defer mt.Unlock()
	return inspect.GetCell(mt.Terminal, uint16(x), uint32(y))
}

// RecentOutput returns the last n bytes of the child's raw output, with
// control bytes rendered visibly so diagnostics stay readable.
func (s *Service) RecentOutput(id string, n int) (string, error) {
	if n <= 0 {
		n = 4096
	}
	if n > 32768 {
		n = 32768
	}
	mt, err := s.terminals.Get(id)
	if err != nil {
		return "", err
	}
	return string(mt.RecentOutput(n)), nil
}
