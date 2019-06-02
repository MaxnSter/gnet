package gnet

import (
	"net"
)

type SessionManager interface {
	// BroadCast对所有NetSession连接执行fn
	// 若module设置Pool,则fn全部投入Pool中,否则在当前goroutine执行
	Broadcast(func(session NetSession))

	// GetSession返回指定id对应的NetSession
	GetSession(id uint64) (NetSession, bool)
}

type NetSession interface {
	ID() uint64
	LocalAddr() net.Addr
	RemoteAddr() net.Addr

	Send(message interface{})
	AccessManager() SessionManager

	Runner
}

type NetServer interface {
	SessionManager
	Runner
}

type NetClient interface {
	SessionManager
	Runner
}
