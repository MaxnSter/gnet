package timer

import (
	"time"

	"github.com/MaxnSter/gnet/pool"
)

// OnTimeOut 是定时器的callback
type OnTimeOut func(time.Time)
type Cancel func()

type Timer interface {
	Run()
	Stop()

	SetPool(pool.Pool)
	AddTimer(expire time.Time, interval time.Duration, cb OnTimeOut) Cancel

	String() string
}
