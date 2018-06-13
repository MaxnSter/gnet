package timer

import (
	"time"

	"github.com/MaxnSter/gnet/gnet_context"
	"github.com/MaxnSter/gnet/worker_pool"
)

type OnTimeOut func(time.Time, gnet_context.Context)

type TimerManager interface {
	Start()
	Stop()
	StopAsync() <-chan struct{}

	AddTimer(expire time.Time, interval time.Duration, ctx gnet_context.Context, cb OnTimeOut) (timerId int64)
	CancelTimer(id int64)
}

func NewTimerManager(pool worker_pool.Pool) TimerManager {
	return newTimerManager(pool)
}