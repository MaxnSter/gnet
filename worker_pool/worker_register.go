package worker_pool

type workerPoolCreator func() Pool

var (
	workerPools = map[string]workerPoolCreator{}
)

func RegisterWorkerPool(name string, creator workerPoolCreator) {
	if _, ok := workerPools[name]; ok {
		panic("dup register Pool, name : " + name)
	}

	workerPools[name] = creator
}

func MustGetWorkerPool(name string) Pool {
	if creator, ok := workerPools[name]; !ok {
		panic("Pool not register, name : " + name)
	} else {
		return creator()
	}
}
