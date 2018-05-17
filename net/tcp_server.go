package net

import (
	"errors"
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/logger"
	"github.com/sirupsen/logrus"
)

type TcpServer struct {
	addr     string
	name     string
	options  *NetOptions
	listener *net.TCPListener

	sessions sync.Map
	idGen    int64

	guard   *sync.Mutex
	wg      *sync.WaitGroup
	started bool
	stopped bool

	stopCh chan error
}

func NewTcpServer(addr, name string, options *NetOptions) *TcpServer {
	return &TcpServer{
		options:  options,
		addr:     addr,
		name:     name,
		sessions: sync.Map{},
		guard:    &sync.Mutex{},
		wg:       &sync.WaitGroup{},
		stopCh:   make(chan error),
	}
}

func (server *TcpServer) Start() error {

	server.guard.Lock()
	if server.started {
		server.guard.Unlock()
		return errors.New("server already started")
	}
	server.started = true
	server.guard.Unlock()

	logger.WithField("addr", server.addr).Infoln("server start listening")
	l, err := net.Listen("tcp", server.addr)
	if err != nil {
		logger.WithFields(logrus.Fields{"addr": server.addr, "error": err}).Errorln("server listen error")

		return err
	}
	server.listener = l.(*net.TCPListener)


	return nil
}

func (server *TcpServer) accept() {
	defer func() {

		if r := recover(); r != nil {
			logger.WithField("error", r).Errorln("acceptor recover from error")
		}

		server.wg.Done()
		logger.Infoln("acceptor stopped")
	}()

	logger.Infoln("server start running finished, waiting for connect...")

	delayTime := 5 * time.Microsecond
	maxDelayTime := time.Second
	for {

		select {
		case <-server.stopCh:
			return
		default:
		}

		conn, err := server.listener.AcceptTCP()
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Temporary() {

				if delayTime*2 >= maxDelayTime {
					delayTime = maxDelayTime
				} else {
					delayTime *= 2
				}

				logger.WithField("error", err).Warningln("acceptor got temporary error")
				time.Sleep(delayTime)
				continue

			} else {
				panic(err)
				//return
			}
		}

		delayTime = 0

		server.wg.Add(1)
		go server.onNewConnection(conn)
		logger.WithField("addr", conn.RemoteAddr().String()).Debugln("new connection accepted")
	}
}

func (server *TcpServer) onNewConnection(conn *net.TCPConn) {
	sid := atomic.AddInt64(&server.idGen, 1)
	session := NewTcpSession(sid, server.options, conn, func(s *TcpSession) {
		//after session close done
		server.sessions.Delete(s.id)
		server.wg.Done()
	})
	session.SetManager(server)
	server.sessions.Store(sid, session)

	session.Start()
}

func (server *TcpServer) Run() {

	server.guard.Lock()
	if !server.started {
		server.guard.Unlock()
		panic("server not started!")
	}
	server.guard.Unlock()

	//TCP_NODELAY  默认true
	//SO_REUSEADDR 默认设置
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	signal.Ignore(syscall.SIGPIPE)

	logger.Infoln("server start running...")

	//开启worker pool
	server.options.Worker.Start()

	//开启timer manager
	server.options.Timer.Start()

	//开始接收连接
	server.wg.Add(1)
	go server.accept()

	//监听信号
	go func() {
		s := <-sigCh
		logger.WithField("signal", s).Infoln("catch signal")

		server.Stop()
	}()

	//等待服务器关闭,等待所有session退出
	server.wg.Wait()
	logger.Infoln("all session closed")

	<-server.options.Worker.Stop()
	<-server.options.Timer.Stop()

	if server.options.OnServerClosed != nil {
		logger.Infoln("server closed, callback to user")

		server.options.OnServerClosed()
	}

	logger.Infoln("server closed, exit...")
}

func (server *TcpServer) Stop() {
	server.guard.Lock()
	if server.stopped {
		server.guard.Unlock()
		return
	}

	server.stopped = true
	server.guard.Unlock()

	logger.Infoln("server start closing...")

	//立即停止接收新连接
	logger.Infoln("closing listener...")
	server.listener.Close()

	//关闭所有在线连接
	logger.Infoln("closing all sessions...")
	server.AccessSession(func(Id, session interface{}) bool {
		session.(iface.NetSession).Stop()
		return true
	})

	close(server.stopCh)
}

func (server *TcpServer) StartAndRun() {
	if err := server.Start(); err != nil {
		panic(err)
	}

	server.Run()
}

func (server *TcpServer) AccessSession(accessFunc func(Id, session interface{}) bool) {
	server.sessions.Range(accessFunc)
}
