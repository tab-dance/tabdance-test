package service

import (
	"fmt"

	"github.com/tab-dance/tabdance-test/internal/input"
)

func (s *Service) SendText(id, text string) error {
	mt, err := s.terminals.Get(id)
	if err != nil {
		return err
	}
	mt.Lock()
	defer mt.Unlock()
	if _, err := mt.WritePty([]byte(text)); err != nil {
		return err
	}
	return nil
}

func (s *Service) SendKeys(id string, keys []input.KeyInput) error {
	mt, err := s.terminals.Get(id)
	if err != nil {
		return err
	}

	parsed := make([]input.KeyInput, len(keys))
	copy(parsed, keys)
	for i, ki := range parsed {
		if ki.Mods == "" && ki.Text == "" {
			parsed[i] = input.ParseKeyCombo(ki.Key)
		}
	}

	mt.Lock()
	defer mt.Unlock()
	if len(parsed) == 0 {
		return fmt.Errorf("no valid keys provided")
	}
	return input.SendKeys(mt.KeyEncoder, mt, mt.Terminal, parsed)
}

func (s *Service) SendMouse(id string, mi input.MouseInput) error {
	mt, err := s.terminals.Get(id)
	if err != nil {
		return err
	}
	mt.Lock()
	defer mt.Unlock()
	if input.IsScrollAction(mi.Action) {
		return input.SendScroll(mt.MouseEncoder, mt, mi.Action, mi.X, mi.Y, mi.Mods)
	}
	if err := input.SendMouse(mt.MouseEncoder, mt, mi); err != nil {
		return err
	}
	return nil
}

func (s *Service) SendFocus(id, event string) error {
	mt, err := s.terminals.Get(id)
	if err != nil {
		return err
	}
	mt.Lock()
	defer mt.Unlock()
	return input.SendFocus(mt, event)
}

func (s *Service) SendPaste(id, text string) error {
	mt, err := s.terminals.Get(id)
	if err != nil {
		return err
	}
	mt.Lock()
	defer mt.Unlock()
	return input.SendPaste(mt, mt.Terminal, text)
}

func (s *Service) Resize(id string, cols, rows int) error {
	mt, err := s.terminals.Get(id)
	if err != nil {
		return err
	}
	if cols <= 0 || rows <= 0 {
		return fmt.Errorf("resize needs positive dimensions, got %dx%d", cols, rows)
	}
	return mt.Resize(uint16(cols), uint16(rows))
}
