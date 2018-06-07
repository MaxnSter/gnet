package iface

type Identifier interface {
	ID() int64
}

type NetSession interface {
	Identifier

	Send(message interface{})
	Stop()

	//LoadCtx(key interface{}) (val interface{}, ok bool)
	//StoreCtx(key interface{}, val interface{})
}

type SessionManager interface {
	AddSession() error
	RmSession(s NetSession) error
}

type Server interface {
	SessionManager

	Start(userCb OnMessage)
	Stop()
}
