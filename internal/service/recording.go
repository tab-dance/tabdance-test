package service

func (s *Service) StartRecording(id, path string, idleTimeLimit float64) error {
	mt, err := s.terminals.Get(id)
	if err != nil {
		return err
	}
	if idleTimeLimit <= 0 {
		idleTimeLimit = 2.0
	}
	return mt.StartRecording(path, idleTimeLimit)
}

func (s *Service) StopRecording(id string) error {
	mt, err := s.terminals.Get(id)
	if err != nil {
		return err
	}
	return mt.StopRecording()
}

func (s *Service) AddMarker(id, label string) error {
	mt, err := s.terminals.Get(id)
	if err != nil {
		return err
	}
	return mt.AddMarker(label)
}
