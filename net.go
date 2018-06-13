package gnet

import (
	"io"

	"github.com/MaxnSter/gnet/iface"
)

type Property interface {
	LoadCtx(key interface{}) (val interface{}, ok bool)
	StoreCtx(key interface{}, val interface{})
}

type SessionManager interface {
	Broadcast(func(session NetSession))
	GetSession(id int64) (NetSession, bool)
}

type NetSession interface {
	iface.Identifier
	Property

	Raw() io.ReadWriter
	Send(message interface{})
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
