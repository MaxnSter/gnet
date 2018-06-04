package iface

type NetSession interface {
	ID() int64
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
