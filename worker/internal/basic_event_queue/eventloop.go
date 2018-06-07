package basic_event_queue

import (
	"errors"
	"runtime/debug"

	"github.com/MaxnSter/gnet/iface"
)

type CallBackWrapper func(ctx iface.Context, cb func(iface.Context))

func Decorate(usrCtx iface.Context, usrCb func(iface.Context), wrapper CallBackWrapper) func() {
	return func() {
		wrapper(usrCtx, usrCb)
	}
}

var (
	SafeCallBack = func(ctx iface.Context, cb func(iface.Context)) {
		defer func() {
			if r := recover(); r != nil {
				// TODO error handing
				debug.PrintStack()
			}
		}()

		cb(ctx)
	}

	UnSafeCallBack = func(ctx iface.Context, cb func(iface.Context)) {
		cb(ctx)
	}
)

type EventQueue struct {
	queue     chan func()
	closeDone chan struct{}

	queueSize      int
	isSafeCallBack bool
	cbWrapper      CallBackWrapper
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
		loop.cbWrapper = SafeCallBack
	} else {
		loop.cbWrapper = UnSafeCallBack
	}

	go loop.loop()
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

func (loop *EventQueue) Put(ctx iface.Context, cb func(iface.Context)) error {

	select {
	case loop.queue <- Decorate(ctx, cb, loop.cbWrapper):
		return nil
	default:
		//TODO error
		return errors.New("queue size limit")
	}
}

func (loop *EventQueue) MustPut(ctx iface.Context, cb func(iface.Context)) {
	loop.queue <- Decorate(ctx, cb, loop.cbWrapper)
}
