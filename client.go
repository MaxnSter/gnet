package gnet

import (
	"github.com/MaxnSter/gnet/pool"
	"github.com/MaxnSter/gnet/util"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type client struct {
	NetSession
	Module
	operator Operator

	sync.Once
}

func NewClient(conn net.Conn, m Module, o Operator) NetClient {
	c := &client{
		Module:   m,
		operator: o,
	}

	id := util.GetUUID()
	c.NetSession = newSession(id, conn, c, o)
	return c
}

func (c *client) Run() {
	c.Once.Do(func() {
		go c.signal()

		c.Pool().Run()
		c.NetSession.Run()
		c.Pool().Stop()
	})
}

func (c *client) signal() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	signal.Ignore(syscall.SIGPIPE)

	<-sigCh
	c.Stop()
}

func (c *client) Broadcast(f func(session NetSession)) {
	c.Pool().Put(func() {
		f(c)
	}, pool.WithIdentify(c))
}

func (c *client) GetSession(id uint64) (NetSession, bool) {
	if c.ID() != id {
		return nil, false
	}
	return c, true
}
