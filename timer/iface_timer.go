package timer

import (
	"time"

	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/worker_pool"
)

type OnTimeOut func(time.Time, iface.Context)

type TimerManager interface {
	Start()
	Stop()
	StopAsync() <-chan struct{}

	AddTimer(expire time.Time, interval time.Duration, ctx iface.Context, cb OnTimeOut) (timerId int64)
	CancelTimer(id int64)
}

func NewTimerManager(pool worker_pool.Pool) TimerManager {
	return newTimerManager(pool)
}