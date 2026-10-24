package terminal

import "sync/atomic"

type outputSubscriber struct {
	ch chan []byte
	// dropped counts bytes lost because the subscriber fell behind; used
	// by attach to report loss instead of failing silently.
	dropped atomic.Int64
}

func (s *outputSubscriber) Channel() <-chan []byte {
	return s.ch
}

func (s *outputSubscriber) Dropped() int64 { return s.dropped.Load() }
