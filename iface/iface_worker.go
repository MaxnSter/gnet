package iface

type WorkerPool interface {
	Start()
	Stop()
	StopAsync() (done <-chan struct{})

	Put(session NetSession, cb func())
	TryPut(session NetSession, cb func()) bool

	TypeName() string
}
