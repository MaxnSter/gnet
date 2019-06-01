package basic_event_queue

import (
	"errors"
	"sync"
)

type EventQueue struct {
	queue     chan func()
	closeDone chan struct{}

	queueSize int
	sync.Once
}

func NewEventQueue(queueSize int) *EventQueue {
	loop := &EventQueue{
		queue:     make(chan func(), queueSize),
		closeDone: make(chan struct{}),
		queueSize: queueSize,
	}

	return loop
}

func (loop *EventQueue) Run() {
	loop.Do(func() {
		go loop.loop()
	})
}

func (loop *EventQueue) loop() {
	for cb := range loop.queue {
		if cb == nil {
			break
		}

		cb()
	}

	if loop.closeDone != nil {
		close(loop.closeDone)
	}
}

func (loop *EventQueue) StopAsync() (done <-chan struct{}) {
	loop.queue <- nil
	return loop.closeDone
}

func (loop *EventQueue) Stop() {
	<-loop.StopAsync()
}

func (loop *EventQueue) Put(f func()) error {
	select {
	case loop.queue <- f:
		return nil
	default:
		return errors.New("queue size limit")
	}
}

func (loop *EventQueue) MustPut(f func()) {
	go func() {
		loop.queue <- f
	}()
}
