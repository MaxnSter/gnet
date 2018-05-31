package iface

import (
	"time"
)

type TimeOutCB func(time.Time, NetSession)

type Timer interface {
	Start()
	Stop()
	StopAsync() <-chan struct{}

	AddTimer(expire time.Time, interval time.Duration, s NetSession, cb TimeOutCB) (timerId int64)
	CancelTimer(id int64)
}
