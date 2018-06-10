package iface

type Identifier interface {
	ID() int64
}

type Property interface {
	LoadCtx(key interface{}) (val interface{}, ok bool)
	StoreCtx(key interface{}, val interface{})
}

type NetSession interface {
	Identifier

	Send(message interface{})
	Stop()
}

type SessionManager interface {
	AddSession() error
	RmSession(s NetSession) error
}

type NetServer interface {
	SessionManager

	Start()
	Stop()
}
