package net

import (
	"errors"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/gnet/util"
	"github.com/sirupsen/logrus"
)

type TcpServer struct {
	addr     string
	name     string
	Options  *NetOptions
	listener *net.TCPListener

	sessions sync.Map

	guard   *sync.Mutex
	wg      *sync.WaitGroup
	started bool
	stopped bool

	stopCh chan error
}

func NewTcpServer(addr, name string, options *NetOptions) *TcpServer {
	return &TcpServer{
		Options:  options,
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
		return errors.New("memcached_server already started")
	}
	server.started = true
	server.guard.Unlock()

	logger.WithField("addr", server.addr).Infoln("memcached_server start listening")
	l, err := net.Listen("tcp", server.addr)
	if err != nil {
		logger.WithFields(logrus.Fields{"addr": server.addr, "error": err}).Errorln("memcached_server listen error")

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

	logger.Infoln("memcached_server start running finished, waiting for connect...")

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
	sid := util.GetUUID()
	session := NewTcpSession(sid, server.Options, conn, func(s *TcpSession) {
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
		panic("memcached_server not started!")
	}
	server.guard.Unlock()

	//TCP_NODELAY  默认true
	//SO_REUSEADDR 默认设置
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	signal.Ignore(syscall.SIGPIPE)

	logger.Infoln("memcached_server start running...")

	//开启worker pool
	server.Options.Worker.Start()

	//开启timer manager
	server.Options.Timer.Start()

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

	<-server.Options.Worker.Stop()
	<-server.Options.Timer.Stop()

	if server.Options.OnServerClosed != nil {
		logger.Infoln("memcached_server closed, callback to user")

		server.Options.OnServerClosed()
	}

	logger.Infoln("memcached_server closed, exit...")
}

func (server *TcpServer) Stop() {
	server.guard.Lock()
	if server.stopped {
		server.guard.Unlock()
		return
	}

	server.stopped = true
	server.guard.Unlock()

	logger.Infoln("memcached_server start closing...")

	//立即停止接收新连接
	logger.Infoln("closing listener...")
	server.listener.Close()

	server.shutAllSessions()

	close(server.stopCh)
}

func (server *TcpServer) shutAllSessions() {

	//关闭所有在线连接
	logger.Infoln("closing all sessions...")
	server.AccessSession(func(Id, session interface{}) bool {
		session.(iface.NetSession).Stop()
		return true
	})
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
