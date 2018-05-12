package net

import (
	"errors"
	"net"
	"sync"
)

type TcpClient struct {
	addr  string
	netOp NetOptions

	raw     net.Conn
	session *TcpSession

	guard   *sync.Mutex
	wg      *sync.WaitGroup
	started bool
	stopped bool
}

func NewTcpClient(addr string, op NetOptions) *TcpClient {
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
	client.guard.Unlock()

	conn, err := net.Dial("tcp", client.addr)
	if err != nil {
		//TODO retry
		return err
	}

	client.raw = conn
	client.wg.Add(1)
	go client.onNewSession(conn)

	client.guard.Lock()
	client.started = true
	client.guard.Unlock()
	return nil
}

func (client *TcpClient) onNewSession(conn net.Conn) {
	s := NewTcpSession(0, &client.netOp, conn, func(s *TcpSession) {
		client.wg.Done()
	})

	s.Start()
}

func (client *TcpClient) Run() {
	client.guard.Lock()
	if !client.started {
		client.guard.Unlock()
		panic("client not started!")
	}
	client.guard.Unlock()

	//wait for session to stop
	client.wg.Wait()
}

func (client *TcpClient) Stop() {
	client.guard.Lock()
	if client.stopped {
		client.guard.Unlock()
		panic("client already stop!")
	}

	client.stopped = true
	client.guard.Unlock()

	//close the session
	client.session.Close()
}

func (client *TcpClient) StartAndRun()  {
	if err := client.Start(); err != nil {
		panic(err)
	}

	client.Run()
}