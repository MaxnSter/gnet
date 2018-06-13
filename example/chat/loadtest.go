package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"

	"github.com/MaxnSter/gnet/example/chat/client"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/gnet/net"
	"github.com/MaxnSter/gnet/worker_pool"
	_ "github.com/MaxnSter/gnet/worker_pool/worker_session_race_other"
)

var (
	addr        = flag.String("addr", ":2007", "server address")
	connections = flag.Int("c", 1, "connection number")
)

func main() {
	flag.Parse()

	go func() {
		if err := http.ListenAndServe(":8088", nil); err != nil {
			panic(err)
		}
	}()
	new(chatLoadTest).startLoadTest()
}

type chatLoadTest struct {
	conns int

	received int
	start    time.Time
}

func (c *chatLoadTest) startLoadTest() {
	wg := sync.WaitGroup{}
	pool := worker_pool.MustGetWorkerPool("poolRaceOther")
	pool.Start()
	for i := 0; i < *connections; i++ {
		wg.Add(1)
		go func() {
			client.NewChatClientWithPool(*addr, c.onConnected, c.onMessage, pool).Run()
			wg.Done()
		}()
	}

	wg.Wait()
}

func (c *chatLoadTest) onConnected(session *net.TcpSession) {
	c.conns++
	if c.conns == *connections {
		logger.Infoln("all connected")
		session.RunAfter(time.Now(), time.Second*5, func(i time.Time, context iface.Context) {
			c.start = time.Now()
			s := context.(iface.NetSession)
			s.Send("hello")
		})
	}
}

func (c *chatLoadTest) onMessage(_ iface.Event) {
	c.received++
	if c.received == c.conns {
		end := time.Now()
		logger.Infoln("all received ", c.conns, " in ", end.Sub(c.start).Seconds())
	}
}
