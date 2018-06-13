package tcp

import (
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/gnet/util"
)

const (
	baseRetryDuration = time.Millisecond * 500
	maxRetryDuration  = time.Second * 3
)

type tcpClient struct {
	addr          string
	name          string
	sessionNumber int
	sessions      sync.Map
	status        *status

	wg       *sync.WaitGroup
	module   gnet.Module
	operator gnet.Operator
}

func newTcpClient(name string, m gnet.Module, o gnet.Operator) gnet.NetClient {
	return &tcpClient{
		wg:            &sync.WaitGroup{},
		status:        &status{},
		name:          name,
		module:        m,
		sessionNumber: 1,
		operator:      o,
	}
}

func (c *tcpClient) SetSessionNumber(sessionNumber int) {
	c.sessionNumber = sessionNumber
}

func (c *tcpClient) Connect(addr string) {
	if !c.status.start() {
		//TODO duplicate Connect()
		return
	}

	logger.WithField("addr", c.addr).Infoln("client connecting to ...")
	c.addr = addr

	c.wg.Add(c.sessionNumber)
	for i := 0; i < c.sessionNumber; i++ {
		go c.connect()
	}

	c.Run()
}

func (c *tcpClient) connect() {
	var conn net.Conn
	var err error
	curRetryDuration := baseRetryDuration

	for {
		conn, err = net.Dial("tcp", c.addr)

		if err == nil {
			logger.WithField("addr", c.addr).Infoln("client connected to ")
			break
		}

		curRetryDuration *= 2
		if curRetryDuration > maxRetryDuration {
			curRetryDuration = maxRetryDuration
		}

		logger.WithField("error", err).Errorln("client connect error, retrying...")
		time.Sleep(curRetryDuration)
	}

	c.onNewSession(conn.(*net.TCPConn))
}

func (c *tcpClient) onNewSession(conn *net.TCPConn) {
	sid := util.GetUUID()
	s := NewTcpSession(sid, conn, c.module, c.operator, func(s *tcpSession) {
		c.sessions.Delete(s.ID())
		c.wg.Done()
	})

	c.sessions.Store(sid, s)
	s.Start()
}

func (c *tcpClient) setSignal() {
	sigCh := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGPIPE)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)

	go func() {
		s := <-sigCh
		logger.WithField("signal", s).Infoln("c catch signal")
		c.Stop()
	}()
}

func (c *tcpClient) Run() {
	c.operator.StartModule(c.module)

	//wait for session to stop
	c.wg.Wait()
	logger.Infoln("all sessions closed")

	//stop module
	c.operator.StopModule(c.module)
	logger.Infoln("client closed, exit...")
}

func (c *tcpClient) Stop() {
	if !c.status.stop() {
		//TODO duplicate stop()
		return
	}

	logger.Infoln("client stopping...")
	logger.Infoln("stopping session...")
	c.Broadcast(func(session gnet.NetSession) {
		session.Stop()
	})
}

func (c *tcpClient) Broadcast(fn func(session gnet.NetSession)) {
	//FIXME callback hell
	c.sessions.Range(func(id, session interface{}) bool {
		c.module.Pool().Put(session, func(ctx iface.Context) {
			fn(ctx.(gnet.NetSession))
		})
		return true
	})
}

func (c *tcpClient) GetSession(id int64) (gnet.NetSession, bool) {
	if session, ok := c.sessions.Load(id); ok {
		return session.(gnet.NetSession), true
	} else {
		return nil, false
	}
}

func init() {
	gnet.RegisterClientCreator("tcp", newTcpClient)
}
