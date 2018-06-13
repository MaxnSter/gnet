package worker_pool

import "github.com/MaxnSter/gnet/gnet_context"

type Pool interface {
	Start()
	Stop()
	StopAsync() (done <-chan struct{})

	Put(ctx gnet_context.Context, cb func(gnet_context.Context))
	TryPut(ctx gnet_context.Context, cb func(ctx gnet_context.Context)) bool

	TypeName() string
}
