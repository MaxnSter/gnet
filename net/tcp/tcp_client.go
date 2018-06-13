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

func (client *tcpClient) SetSessionNumber(sessionNumber int) {
	client.sessionNumber = sessionNumber
}

func (client *tcpClient) Connect(addr string) {
	if !client.status.start() {
		//TODO duplicate Connect()
		return
	}

	logger.WithField("addr", client.addr).Infoln("client connecting to ...")
	client.addr = addr

	client.wg.Add(client.sessionNumber)
	for i := 0; i < client.sessionNumber; i++ {
		go client.connect()
	}

	client.Run()
}

func (client *tcpClient) connect() {
	var conn net.Conn
	var err error
	curRetryDuration := baseRetryDuration

	for {
		conn, err = net.Dial("tcp", client.addr)

		if err == nil {
			logger.WithField("addr", client.addr).Infoln("client connected to ")
			break
		}

		curRetryDuration *= 2
		if curRetryDuration > maxRetryDuration {
			curRetryDuration = maxRetryDuration
		}

		logger.WithField("error", err).Errorln("client connect error, retrying...")
		time.Sleep(curRetryDuration)
	}

	client.onNewSession(conn.(*net.TCPConn))
}

func (client *tcpClient) onNewSession(conn *net.TCPConn) {
	sid := util.GetUUID()
	s := NewTcpSession(sid, conn, client.module, client.operator, func(s *tcpSession) {
		client.sessions.Delete(s.ID())
		client.wg.Done()
	})

	client.sessions.Store(sid, s)
	s.Start()
}

func (client *tcpClient) setSignal() {
	sigCh := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGPIPE)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)

	go func() {
		s := <-sigCh
		logger.WithField("signal", s).Infoln("client catch signal")
		client.Stop()
	}()
}

func (client *tcpClient) Run() {
	client.operator.StartModule(client.module)

	//wait for session to stop
	client.wg.Wait()
	logger.Infoln("all sessions closed")

	//stop module
	client.operator.StopModule(client.module)
	logger.Infoln("client closed, exit...")
}

func (client *tcpClient) Stop() {
	if !client.status.stop() {
		//TODO duplicate stop()
		return
	}

	logger.Infoln("client stopping...")
	logger.Infoln("stopping session...")
	client.Broadcast(func(session gnet.NetSession) {
		session.Stop()
	})
}

func (client *tcpClient) Broadcast(fn func(session gnet.NetSession)) {
	//FIXME callback hell
	client.sessions.Range(func(id, session interface{}) bool {
		client.module.Pool().Put(session, func(ctx iface.Context) {
			fn(ctx.(gnet.NetSession))
		})
		return true
	})
}

func init() {
	gnet.RegisterClientCreator("tcp", newTcpClient)
}
