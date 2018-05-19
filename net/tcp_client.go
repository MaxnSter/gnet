package net

import (
	"errors"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/MaxnSter/gnet/logger"
)

const (
	baseRetryDuraion = time.Millisecond * 500
	maxRetryDuation  = time.Second * 5
)

type TcpClient struct {
	addr  string
	netOp *NetOptions

	raw     net.Conn
	session *TcpSession

	guard   *sync.Mutex
	wg      *sync.WaitGroup
	started bool
	stopped bool
}

func NewTcpClient(addr string, op *NetOptions) *TcpClient {
	return &TcpClient{
		addr:  addr,
		netOp: op,
		guard: &sync.Mutex{},
		wg:    &sync.WaitGroup{},
	}
}

func (client *TcpClient) Start() error {
	client.guard.Lock()
	if client.started {
		client.guard.Unlock()
		return errors.New("client already started")
	}
	client.started = true
	client.guard.Unlock()

	logger.WithField("addr", client.addr).Infoln("client connecting to server...")

	var conn net.Conn
	var err error
	curRetryDuration := baseRetryDuraion

	for curRetryDuration < maxRetryDuation {
		conn, err = net.Dial("tcp", client.addr)

		if err == nil {
			logger.WithField("addr", client.addr).Infoln("client connected to server")
			break
		}

		logger.WithField("error", err).Errorln("client connect error, retrying...")
		curRetryDuration *= 2
	}

	//still can't connect to server
	if err != nil {
		panic(err)
	}

	client.raw = conn
	return nil
}

func (client *TcpClient) onNewSession(conn *net.TCPConn) {
	s := NewTcpSession(0, client.netOp, conn, func(s *TcpSession) {
		client.wg.Done()
	})

	client.session = s
	s.Start()
}

func (client *TcpClient) Run() {
	client.guard.Lock()
	if !client.started {
		client.guard.Unlock()
		panic("client not started!")
	}
	client.guard.Unlock()

	//start worker pool
	client.netOp.Worker.Start()

	//start timer
	client.netOp.Timer.Start()

	client.wg.Add(1)
	go client.onNewSession(client.raw.(*net.TCPConn))

	//忽略SIGPIPE
	//TCP_NODELAY默认开启
	sigCh := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGPIPE)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)

	go func() {
		s := <-sigCh
		logger.WithField("signal", s).Infoln("client catch signal")
		client.Stop()
	}()

	logger.Infoln("client start finished, waiting for event...")

	//wait for session to stop
	client.wg.Wait()
	logger.Infoln("session closed")

	//close worker pool wait close Done
	<-client.netOp.Worker.Stop()

	//close timer and wait for close Done
	<-client.netOp.Timer.Stop()

	logger.Infoln("client closed, exit...")
}

func (client *TcpClient) Stop() {
	client.guard.Lock()
	if client.stopped {
		client.guard.Unlock()
		panic("client already stop!")
	}

	client.stopped = true
	client.guard.Unlock()

	logger.Infoln("client stopping...")

	//close the session
	logger.Infoln("stopping session...")
	client.session.Stop()
}

func (client *TcpClient) StartAndRun() {
	if err := client.Start(); err != nil {
		panic(err)
	}

	client.Run()
}
