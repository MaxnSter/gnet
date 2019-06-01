package tcp

import "sync/atomic"

const (
	ready = iota
	start
	stop
)

type status struct {
	curStatus int64
}

func (s *status) start() bool {
	if !atomic.CompareAndSwapInt64(&s.curStatus, ready, start) {
		return false
	}
	return true
}

func (s *status) stop() bool {
	if !atomic.CompareAndSwapInt64(&s.curStatus, start, stop) {
		return false
	}
	return true
}
