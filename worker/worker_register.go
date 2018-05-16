package worker

import "github.com/MaxnSter/gnet/iface"

type workerPoolCreator func() iface.WorkerPool

var (
	workerPools = map[string]workerPoolCreator{}
)

func RegisterWorkerPool(name string, creator workerPoolCreator) {
	if _, ok := workerPools[name]; ok {
		panic("dup register WorkerPool, name : " + name)
	}

	workerPools[name] = creator
}

func MustGetWorkerPool(name string) iface.WorkerPool {
	if creator, ok := workerPools[name]; !ok {
		panic("WorkerPool not register, name : " + name)
	} else {
		return creator()
	}
}
