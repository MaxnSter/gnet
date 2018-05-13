package iface

type NetSession interface {
	ID() int64

	Send(message Message)

	//LoadCtx(key interface{}) (val interface{})
	//StoreCtx(key interface{}, val interface{})
}

type SessionManager interface {
	AddSession() error
	RmSession(s NetSession) error
}

type Server interface {
	SessionManager

	Start(userCb UserEventCB)
	Stop()
}
