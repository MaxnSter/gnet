package iface

type WorkerPool interface {
	Start()
	Stop() (done <-chan struct{})

	//TODO
	Put(session NetSession, cb func())

	TypeName() string
}
