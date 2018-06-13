package worker_pool

import "github.com/MaxnSter/gnet/iface"

type Pool interface {
	Start()
	Stop()
	StopAsync() (done <-chan struct{})

	Put(ctx iface.Context, cb func(iface.Context))
	TryPut(ctx iface.Context, cb func(ctx iface.Context)) bool

	TypeName() string
}
