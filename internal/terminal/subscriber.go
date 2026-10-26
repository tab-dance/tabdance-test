package terminal

import "sync"

type outputSubscriber struct {
	ch   chan []byte
	quit chan struct{}
	// mu serializes send vs close: deliver holds it for the duration of a
	// blocking send, and closeUnsubscribed only closes after quit was
	// signaled, so a sender can never touch the channel after it is closed.
	mu     sync.Mutex
	closed bool
}

func newOutputSubscriber(queueSize int) *outputSubscriber {
	return &outputSubscriber{
		ch:   make(chan []byte, queueSize),
		quit: make(chan struct{}),
	}
}

func (s *outputSubscriber) Channel() <-chan []byte {
	return s.ch
}

// deliver enqueues chunk for the consumer, blocking the PTY read loop until
// there is room. Raw output chunks cannot be dropped without corrupting the
// escape sequence stream, so a slow consumer back-pressures the child
// instead, exactly like a real terminal would.
func (s *outputSubscriber) deliver(chunk []byte) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return false
	}
	select {
	case s.ch <- chunk:
		return true
	case <-s.quit:
		return false
	}
}

// closeUnsubscribed terminates the consumer's range loop. Call only after
// the subscriber was removed from the broadcast list so no new deliveries
// can start; close(quit) then wakes a pending blocking send.
func (s *outputSubscriber) closeUnsubscribed() {
	close(s.quit)
	s.mu.Lock()
	if !s.closed {
		s.closed = true
		close(s.ch)
	}
	s.mu.Unlock()
}
