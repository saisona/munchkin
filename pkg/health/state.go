package health

import "sync/atomic"

type StartupState struct {
	ready atomic.Bool
}

func NewStartupState() *StartupState {
	return &StartupState{}
}

func (s *StartupState) MarkReady() {
	s.ready.Store(true)
}

func (s *StartupState) IsReady() bool {
	return s.ready.Load()
}

