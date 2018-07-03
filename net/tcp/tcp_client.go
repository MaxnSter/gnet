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
	sessionNumber int //同时建立几个连接
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

// SetSessionNumber设置客户端连接数,默认为1
func (c *tcpClient) SetSessionNumber(sessionNumber int) {
	if sessionNumber < 1 {
		c.sessionNumber = 1
		return
	}
	c.sessionNumber = sessionNumber
}

// Connect开始建立连接并启动客户端,本次调用将阻塞直到客户端退出
// Connect会自动重试直到成功建立连接
func (c *tcpClient) Connect(addr string) {
	if !c.status.start() {
		return
	}

	logger.WithField("addr", c.addr).Infoln("client connecting to ...")
	c.addr = addr

	c.wg.Add(c.sessionNumber)
	for i := 0; i < c.sessionNumber; i++ {
		go c.connect()
	}

	c.run()
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
	s := newTCPSession(sid, conn, c, c.module, c.operator, func(s *tcpSession) {
		c.sessions.Delete(s.ID())
		c.wg.Done()
	})

	c.sessions.Store(sid, s)
	s.start()
}

func (c *tcpClient) setSignal() {
	sigCh := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGPIPE)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)

	go func() {
		s := <-sigCh
		logger.WithField("signal", s).Infoln("catch signal")
		c.Stop()
	}()
}

func (c *tcpClient) run() {
	c.operator.StartModule(c.module)

	//wait for session to stop
	c.wg.Wait()
	logger.Infoln("all sessions closed")

	//stop module
	c.operator.StopModule(c.module)
	logger.Infoln("client closed, exit...")
}

// Stop停止客户端,关闭当前所有客户端连接
func (c *tcpClient) Stop() {
	if !c.status.stop() {
		return
	}

	logger.Infoln("client stopping...")
	logger.Infoln("stopping session...")
	c.Broadcast(func(session gnet.NetSession) {
		session.Stop()
	})
}

// BroadCast对所有NetSession连接执行fn
// 若module设置Pool,则fn全部投入Pool中,否则在当前goroutine执行
func (c *tcpClient) Broadcast(fn func(session gnet.NetSession)) {
	if c.module.Pool() == nil {
		c.sessions.Range(func(id, session interface{}) bool {
			fn(session.(gnet.NetSession))
			return true
		})
		return
	}

	//FIXME callback hell
	c.sessions.Range(func(id, session interface{}) bool {
		c.module.Pool().Put(session, func(ctx iface.Context) {
			fn(ctx.(gnet.NetSession))
		})
		return true
	})
}

// GetSession返回指定id对应的NetSession
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
