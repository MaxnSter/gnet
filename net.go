package gnet

import (
	"io"
	"time"

	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/timer"
)

type Property interface {
	LoadCtx(key interface{}) (val interface{}, ok bool)
	StoreCtx(key interface{}, val interface{})
}

type SessionManager interface {
	Broadcast(func(session NetSession))
	GetSession(id int64) (NetSession, bool)
}

type ModuleRunner interface {
	RunInPool(func(NetSession))

	RunAt(at time.Time, cb timer.OnTimeOut) (timerId int64)
	RunAfter(after time.Duration, cb timer.OnTimeOut) (timerId int64)
	RunEvery(at time.Time, interval time.Duration, cb timer.OnTimeOut) (timerId int64)
	CancelTimer(timeId int64)
}

type NetSession interface {
	iface.Identifier
	Property
	ModuleRunner

	Raw() io.ReadWriter
	Send(message interface{})
	AccessManager() SessionManager
	Stop()
}

type NetServer interface {
	SessionManager

	Serve()
	Listen(addr string) error
	ListenAndServe(addr string)

	Stop()
}

type NetClient interface {
	SessionManager
	SetSessionNumber(sessionNumber int)

	Connect(addr string)
	Stop()
}
