package iface

type WorkerPool interface {
	Start()
	Stop() (done <-chan struct{})

	//TODO
	Put(session NetSession, cb func())
	TryPut(session NetSession, cb func()) bool

	TypeName() string
}
