package timer

import (
	"time"

	"github.com/MaxnSter/gnet/pool"
)

// OnTimeOut 是定时器的callback
type OnTimeOut func(time.Time)
type Cancel func()

type Manager interface {
	Run()
	Stop()

	// AddTimer添加一个定时任务,并返回该任务对应的id
	AddTimer(expire time.Time, interval time.Duration, cb OnTimeOut) Cancel
}

// NewWithPool 通过指定的pool创建一个定时器组件
// pool用于派发所有OnTimeOut回调
func NewWithPool(pool pool.Pool) Manager {
	return newTimerManager(pool)
}

func New() Manager {
	return NewWithPool(nil)
}
