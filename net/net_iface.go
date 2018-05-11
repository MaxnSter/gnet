package net

type Session interface {
	ID() int64
	SetID(id int64)

	LoadCtx(key interface{}) (val interface{})
	StoreCtx(key interface{}, val interface{})
}

type SessionManager interface {
	AddSession() error
	RmSession(s Session) error
}

type Server interface {
	SessionManager

	Start(userCb UserEventCB)
	Stop()
}
