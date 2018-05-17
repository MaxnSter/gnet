package basic_event_queue

import (
	"errors"
	"fmt"
	"runtime/debug"
)

var (
	safeCallBack = func(cb func()) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("catch expected error : %s\n", r)
				debug.PrintStack()
			}
		}()

		cb()
	}

	unSafeCallBack = func(cb func()) {
		cb()
	}
)

type EventQueue struct {
	queue     chan func()
	closeDone chan struct{}

	queueSize      int
	isSafeCallBack bool
	cbWrapper      func(cb func())
}

func NewEventQueue(queueSize int, isSafeCallBack bool) *EventQueue {
	return &EventQueue{
		queue:          make(chan func(), queueSize),
		closeDone:      make(chan struct{}),
		queueSize:      queueSize,
		isSafeCallBack: isSafeCallBack,
	}
}

func (loop *EventQueue) Start() {
	if loop.isSafeCallBack {
		loop.cbWrapper = safeCallBack
	} else {
		loop.cbWrapper = unSafeCallBack
	}

	go loop.loop()
}

func (loop *EventQueue) loop() {

	for cb := range loop.queue {
		if cb == nil {
			break
		}

		loop.cbWrapper(cb)
	}

	if loop.closeDone != nil {
		close(loop.closeDone)
	}
}

func (loop *EventQueue) Stop() (done <-chan struct{}) {
	loop.queue <- nil
	return loop.closeDone
}

func (loop *EventQueue) Put(cb func()) error {
	select {
	case loop.queue <- cb:
		return nil
	default:
		//TODO error
		return errors.New("queue size limit")
	}
}

func (loop *EventQueue) MustPut(cb func()) {
	loop.queue <- cb
}
