package timer

import (
	"time"

	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/worker_pool"
)

// OnTimeOut是定时器的callback
type OnTimeOut func(time.Time, iface.Context)

type TimerManager interface {
	// Start开启定时器功能,此方法保证goroutine safe
	Start()
	// Stop关闭定时器,调用方阻塞直到定时器完全关闭
	Stop()
	// StopAsync关闭定时器,当定时器内部完全关闭时 done可读
	StopAsync() <-chan struct{}

	// AddTimer添加一个定时任务,并返回该任务对应的id
	AddTimer(expire time.Time, interval time.Duration, ctx iface.Context, cb OnTimeOut) (timerId int64)
	// CancelTimer取消一个定时任务
	CancelTimer(id int64)
}

// NewTimerManager通过指定的pool创建一个定时器组件
// pool用于派发所有OnTimeOut回调
//TODO 当pool为nil,使用goroutine派发回调
func NewTimerManager(pool worker_pool.Pool) TimerManager {
	return newTimerManager(pool)
}
