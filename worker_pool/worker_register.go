package worker_pool

type workerPoolCreator func() Pool

var (
	workerPools = map[string]workerPoolCreator{}
)

// RegisterWorkerPool 注册一个pool.
// 如果name已存在,则panic
func RegisterWorkerPool(name string, creator workerPoolCreator) {
	if _, ok := workerPools[name]; ok {
		panic("dup register Pool, name : " + name)
	}

	workerPools[name] = creator
}

// MustGetPacker 获取指定名字对应的pool.
// 若未注册,则panic
func MustGetWorkerPool(name string) Pool {
	if creator, ok := workerPools[name]; !ok {
		panic("Pool not register, name : " + name)
	} else {
		return creator()
	}
}
