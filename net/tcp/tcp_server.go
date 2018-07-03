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

type tcpServer struct {
	name     string
	addr     string
	listener *net.TCPListener
	status   *status

	sessions sync.Map
	wg       *sync.WaitGroup
	stopCh   chan error
	module   gnet.Module
	operator gnet.Operator
}

func newTcpServer(name string, module gnet.Module, operator gnet.Operator) gnet.NetServer {
	return &tcpServer{
		module:   module,
		operator: operator,
		name:     name,
		sessions: sync.Map{},
		wg:       &sync.WaitGroup{},
		status:   &status{},
		stopCh:   make(chan error),
	}
}

// Listen开始监听指定的addr
func (s *tcpServer) Listen(addr string) error {
	s.addr = addr
	logger.WithField("addr", s.addr).Infoln("s start listening")

	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	//TCP_NODELAY  默认true
	//SO_REUSEADDR 默认设置
	s.listener = l.(*net.TCPListener)
	return nil
}

func (s *tcpServer) accept() {
	defer func() {
		if r := recover(); r != nil {
			logger.WithField("error", r).Errorln("acceptor recover from error")
		}

		s.wg.Done()
		logger.Infoln("acceptor stopped")
	}()

	logger.Infoln("server start running, waiting for connection...")

	var tempDelay time.Duration
	for {
		select {
		case <-s.stopCh:
			return
		default:
		}

		//TODO 并发数控制
		conn, err := s.listener.AcceptTCP()
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}

				logger.WithField("error", err).Warningln("acceptor got temporary error")
				time.Sleep(tempDelay)
				continue
			} else {
				panic(err)
			}
		}

		tempDelay = 0
		s.wg.Add(1)
		go s.onNewConnection(conn)
	}
}

func (s *tcpServer) onNewConnection(conn *net.TCPConn) {
	logger.WithField("addr", conn.RemoteAddr().String()).Debugln("new connection accepted")

	sid := util.GetUUID()
	session := newTCPSession(sid, conn, s, s.module, s.operator, func(session *tcpSession) {
		//after session close done
		s.sessions.Delete(session.ID())
		s.wg.Done()
	})
	s.sessions.Store(sid, session)

	session.start()
}

// ListenAndServe监听指定addr并启动服务器
// 若Listen成功,则阻塞直到服务器关闭完成
// 若Listen失败,则panic
func (s *tcpServer) ListenAndServe(addr string) {
	if err := s.Listen(addr); err != nil {
		panic(err)
	}
	s.Serve()
}

func (s *tcpServer) setSignal() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	signal.Ignore(syscall.SIGPIPE)

	//监听信号
	go func() {
		sig := <-sigCh
		logger.WithField("signal", sig).Infoln("catch signal")

		s.Stop()
	}()
}

// Serve启动服务器,调用方阻塞直到服务器关闭完成
// Serve必须在Listen成功后才可调用
func (s *tcpServer) Serve() {
	s.setSignal()
	s.operator.StartModule(s.module)

	//开始接收连接
	s.wg.Add(1)
	go s.accept()

	//等待服务器关闭,等待所有session退出
	s.wg.Wait()
	logger.Infoln("all session closed")

	//关闭所有组件
	s.operator.StopModule(s.module)
	logger.Infoln("server closed, exit...")
}

// Stop停止服务器
func (s *tcpServer) Stop() {
	logger.Infoln("server start closing...")

	//立即停止接收新连接
	logger.Infoln("closing listener...")
	s.listener.Close()

	s.shutAllSessions()

	close(s.stopCh)
}

func (s *tcpServer) shutAllSessions() {
	//关闭所有在线连接
	//FIXME any better way?
	logger.Infoln("closing all sessions...")
	s.Broadcast(func(session gnet.NetSession) {
		session.Stop()
	})
}

// BroadCast对所有NetSession连接执行fn
// 若module设置Pool,则fn全部投入Pool中,否则在当前goroutine执行
func (s *tcpServer) Broadcast(fn func(session gnet.NetSession)) {
	if s.module.Pool() == nil {
		s.sessions.Range(func(id, session interface{}) bool {
			fn(session.(gnet.NetSession))
			return true
		})
		return
	}

	//FIXME callback hell
	s.sessions.Range(func(id, session interface{}) bool {
		s.module.Pool().Put(session, func(ctx iface.Context) {
			fn(ctx.(gnet.NetSession))
		})
		return true
	})
}

// GetSession返回指定id对应的NetSession
func (s *tcpServer) GetSession(id int64) (gnet.NetSession, bool) {
	if session, ok := s.sessions.Load(id); ok {
		return session.(gnet.NetSession), true
	} else {
		return nil, false
	}
}

func init() {
	gnet.RegisterServerCreator("tcp", newTcpServer)
}
