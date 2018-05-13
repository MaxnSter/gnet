package net

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type TcpServer struct {
	options  NetOptions
	addr     string
	name     string
	listener net.Listener

	sessions sync.Map
	idGen    int64

	guard   *sync.Mutex
	wg      *sync.WaitGroup
	started bool
	stopped bool

	stopCh chan struct{}
}

func NewTcpServer(addr, name string, options NetOptions) *TcpServer {
	return &TcpServer{
		options:  options,
		addr:     addr,
		name:     name,
		sessions: sync.Map{},
		guard:    &sync.Mutex{},
		wg:       &sync.WaitGroup{},
		stopCh:   make(chan struct{}),
	}
}

func (server *TcpServer) Start() error {

	server.guard.Lock()
	if server.started {
		server.guard.Unlock()
		return errors.New("server already started")
	}
	server.guard.Unlock()

	l, err := net.Listen("tcp", server.addr)
	if err != nil {
		return err
	}

	//start accept
	server.listener = l
	server.wg.Add(1)
	go server.accept()

	//start worker pool
	server.options.Worker.Start()

	server.guard.Lock()
	server.started = true
	server.guard.Unlock()
	return nil
}

func (server *TcpServer) accept() {
	defer func() {

		if r := recover(); r != nil {
			//TODO log err
		}

		//TODO tell caller we are stopping
		server.listener.Close()
		server.wg.Done()
	}()

	delayTime := 5 * time.Microsecond
	maxDelayTime := time.Second
	for {

		select {
		case <-server.stopCh:

			return
		default:
		}

		conn, err := server.listener.Accept()
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Temporary() {

				if delayTime*2 >= maxDelayTime {
					delayTime = maxDelayTime
				} else {
					delayTime *= 2
				}

				time.Sleep(delayTime)
				continue

			} else {
				//TODO
				panic(err)
			}
		}

		delayTime = 0

		server.wg.Add(1)
		go server.onNewConnection(conn)
	}
}

func (server *TcpServer) onNewConnection(conn net.Conn) {
	sid := atomic.AddInt64(&server.idGen, 1)
	session := NewTcpSession(sid, &server.options, conn, func(s *TcpSession) {
		//after session close done
		server.sessions.Delete(s.id)
		server.wg.Done()
	})
	session.SetManager(server)
	server.sessions.Store(sid, session)

	if server.options.OnAccepted != nil {
		server.options.OnAccepted(session)
	}
	session.Start()
}

func (server *TcpServer) Run() {

	server.guard.Lock()
	if !server.started {
		server.guard.Unlock()
		panic("server not started!")
	}
	server.guard.Unlock()

	//wait for listener and all session close
	server.wg.Wait()

	//close workerPool and wait for close Done
	<-server.options.Worker.Stop()

	//TODO release all resource
	server.guard.Lock()
	server.started = false
	server.stopped = false
	server.stopCh = make(chan struct{})
	server.guard.Unlock()

	if server.options.OnServerClose != nil {
		server.options.OnServerClose()
	}
}

func (server *TcpServer) Stop() {
	server.guard.Lock()
	if server.stopped {
		server.guard.Unlock()
		panic("server already stop!")
	}

	server.stopped = true
	server.guard.Unlock()

	close(server.stopCh)
}

func (server *TcpServer) StartAndRun() {
	if err := server.Start(); err != nil {
		panic(err)
	}

	server.Run()
}
