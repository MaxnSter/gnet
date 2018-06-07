package iface

import (
	"time"
)

type OnTimeOut func(time.Time, Context)

type Timer interface {
	Start()
	Stop()
	StopAsync() <-chan struct{}

	AddTimer(expire time.Time, interval time.Duration, ctx Context, cb OnTimeOut) (timerId int64)
	CancelTimer(id int64)
}
