package worker_pool

type Pool interface {
	Start()
	Stop()
	StopAsync() (done <-chan struct{})

	Put(ctx Context, cb func(Context))
	TryPut(ctx Context, cb func(Context)) bool

	TypeName() string
}
