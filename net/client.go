package net

import (
	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/pool"
	"github.com/MaxnSter/gnet/util"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type client struct {
	gnet.NetSession
	gnet.Module
	gnet.Operator

	sync.Once
}

func NewClient(conn net.Conn, m gnet.Module, o gnet.Operator) gnet.NetClient {
	c := &client{
		Module:   m,
		Operator: o,
	}

	id := util.GetUUID()
	c.NetSession = newSession(id, conn, c)
	return c
}

func (c *client) Run() {
	c.Once.Do(func() {
		go c.signal()

		c.NetSession.Run()
	})
}

func (c *client) signal() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	signal.Ignore(syscall.SIGPIPE)

	<-sigCh
	c.Stop()
}

func (c *client) Broadcast(f func(session gnet.NetSession)) {
	c.Pool().Put(func() {
		f(c)
	}, pool.WithIdentify(c))
}

func (c *client) GetSession(id uint64) (gnet.NetSession, bool) {
	if c.ID() != id {
		return nil, false
	}
	return c, true
}
