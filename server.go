package gnet

import (
	"github.com/MaxnSter/gnet/pool"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/MaxnSter/GolangDataStructure/try"
	"github.com/MaxnSter/gnet/util"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type server struct {
	net.Listener
	Module
	operator Operator

	guard    sync.Mutex
	sessions map[uint64]NetSession

	wg   sync.WaitGroup
	once sync.Once
	done chan struct{}
}

func NewServer(l net.Listener, m Module, o Operator) NetServer {
	s := &server{
		Listener: l,
		Module:   m,
		operator: o,
		sessions: map[uint64]NetSession{},
		done:     make(chan struct{}),
	}
	return s
}

func (svc *server) Broadcast(f func(session NetSession)) {
	svc.guard.Lock()
	snapshot := svc.sessions
	svc.guard.Unlock()

	for _, s := range snapshot {
		svc.Pool().Put(func() {
			f(s)
		}, pool.WithIdentify(s))
	}
}

func (svc *server) GetSession(id uint64) (session NetSession, ok bool) {
	svc.guard.Lock()
	defer svc.guard.Unlock()

	session, ok = svc.sessions[id]
	return
}

func (svc *server) Run() {
	svc.once.Do(func() {
		go svc.signal()

		svc.Module.Pool().Run()
		svc.serve()

		svc.wg.Wait()
		svc.Module.Pool().Stop()
	})
}

func (svc *server) signal() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	signal.Ignore(syscall.SIGPIPE)

	<-sigCh
	svc.Stop()
}

func (svc *server) serve() {
	try.Try(func() error {
		var tempDelay time.Duration
		for {
			conn, err := svc.Accept()
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

					time.Sleep(tempDelay)
					continue
				}

				select {
				case <-svc.done:
					return nil
				}
				return errors.Wrap(err, "accept failed")
			}

			go svc.onNewSession(conn)
		}
	}).Final(func(e error) error {
		if e != nil {
			glog.Errorf("%+v", e)
		}

		svc.Stop()
		return nil
	}).Do()
}

func (svc *server) onNewSession(conn net.Conn) {
	id := util.GetUUID()
	session := newSession(id, conn, svc, svc.operator)

	svc.guard.Lock()
	svc.sessions[id] = session
	svc.guard.Unlock()

	svc.wg.Add(1)
	defer func() {
		svc.guard.Lock()
		delete(svc.sessions, id)
		svc.guard.Unlock()

		svc.wg.Done()
	}()
	session.Run()
}

func (svc *server) Stop() {
	select {
	case <-svc.done:
		return
	default:
	}
	close(svc.done)

	svc.Listener.Close()
	svc.Broadcast(func(session NetSession) {
		session.Stop()
	})
}
