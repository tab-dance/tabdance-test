package terminal

type outputSubscriber struct {
	ch chan []byte
}

func (s *outputSubscriber) Channel() <-chan []byte {
	return s.ch
}
