package basic_event_queue

import (
	"errors"
	"runtime/debug"
	"sync"

	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/logger"
)

type CallBackWrapper func(ctx iface.Context, cb func(iface.Context))

func Bind(usrCtx iface.Context, usrCb func(iface.Context), wrapper CallBackWrapper) func() {
	return func() { wrapper(usrCtx, usrCb) }
}

var (
	SafeCallBack = func(ctx iface.Context, cb func(iface.Context)) {
		defer func() {
			if r := recover(); r != nil {
				// TODO error handing
				logger.Errorf("error:%s\n", r)
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

	sync.Once
}

func NewEventQueue(queueSize int, isSafeCallBack bool) *EventQueue {
	loop := &EventQueue{
		queue:          make(chan func(), queueSize),
		closeDone:      make(chan struct{}),
		queueSize:      queueSize,
		isSafeCallBack: isSafeCallBack,
	}

	if loop.isSafeCallBack {
		loop.cbWrapper = SafeCallBack
	} else {
		loop.cbWrapper = UnSafeCallBack
	}

	return loop
}

func (loop *EventQueue) Start() {
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

func (loop *EventQueue) Put(ctx iface.Context, cb func(iface.Context)) error {
	select {
	case loop.queue <- Bind(ctx, cb, loop.cbWrapper):
		return nil
	default:
		//TODO error
		return errors.New("queue size limit")
	}
}

func (loop *EventQueue) MustPut(ctx iface.Context, cb func(iface.Context)) {
	// assert not in loop.loop() goroutine
	// 防止雪崩
	go func() {
		loop.queue <- Bind(ctx, cb, loop.cbWrapper)
	}()
	//loop.queue <- Bind(ctx, cb, loop.cbWrapper)
}
