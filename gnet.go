package gnet

import (
	"github.com/MaxnSter/gnet/codec"
	gnet "github.com/MaxnSter/gnet/net"
	"github.com/MaxnSter/gnet/packer"
	"github.com/MaxnSter/gnet/pool"
	"net"
)

func NewGnetServer(l net.Listener, m Module, o Operator) NetServer {
	return gnet.NewServer(l, m, o)
}

func NewGnetClient(conn net.Conn, m Module, o Operator) NetClient {
	return gnet.NewClient(conn, m, o)
}

func NewModule(pool pool.Pool, c codec.Coder, packer packer.Packer) Module {
	return &moduleWrapper{
		pool:   pool,
		coder:  c,
		packer: packer,
	}
}

func NewOperator(m Module, cb Callback, opts ...func(Operator)) Operator {
	s := &operatorWrapper{
		Module:   m,
		Callback: cb,
	}

	for _, f := range opts {
		f(s)
	}
	return s
}

func WithReadHooks(i ReadInterceptor) func(Operator) {
	return func(operator Operator) {
		operator.(*operatorWrapper).ReadInterceptor = i
	}
}

func WithWriteHooks(i WriteInterceptor) func(Operator) {
	return func(operator Operator) {
		operator.(*operatorWrapper).WriteInterceptor = i
	}
}

