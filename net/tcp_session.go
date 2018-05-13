package net

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/MaxnSter/gnet/iface"
)

type TcpSession struct {
	id      int64
	manager *TcpServer
	netOp   *NetOptions

	guard   *sync.Mutex
	started bool
	closed  bool

	closeCh chan struct{}
	wg      *sync.WaitGroup
	sendCh  chan iface.Message
	raw     net.Conn
	ctx     sync.Map

	onCloseDone func(*TcpSession)
}

func NewTcpSession(id int64, netOp *NetOptions, conn net.Conn, onCloseDone func(*TcpSession)) *TcpSession {
	return &TcpSession{
		id:          id,
		netOp:       netOp,
		raw:         conn,
		onCloseDone: onCloseDone,
		closeCh:     make(chan struct{}),
		sendCh:      make(chan iface.Message, 100), // TODO
		wg:          &sync.WaitGroup{},
		guard:       &sync.Mutex{},
	}
}

func (s *TcpSession) SetManager(m *TcpServer) {
	s.manager = m
}

func (s *TcpSession) ID() int64 {
	return s.id
}

func (s *TcpSession) Start() {
	s.guard.Lock()
	if s.started {
		s.guard.Unlock()
		return
	}
	s.started = true
	s.guard.Unlock()

	//start loop
	s.wg.Add(2)
	go s.readLoop()
	go s.writeLoop()

	//callback to user
	if s.netOp.OnConnect != nil {
		s.netOp.OnConnect(s)
	}

	//wait for session close Done
	s.wg.Wait()
	s.raw.Close()
	if s.onCloseDone != nil {
		s.onCloseDone(s)
	}

}

func (s *TcpSession) Close() {
	s.guard.Lock()
	if s.closed {
		s.guard.Unlock()
		return
	}
	s.closed = true
	s.guard.Unlock()

	if s.netOp.OnClose != nil {
		s.netOp.OnClose(s)
	}

	//close readLoop
	close(s.closeCh)

	//close writeLoop
	s.sendCh <- nil
}

func (s *TcpSession) readLoop() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("catch unexpected error:%s", r)
			s.Close()
		}

		s.wg.Done()
	}()

	for {
		select {
		case <-s.closeCh:
		default:
		}

		msg, err := s.netOp.ReadMessage(s.raw)
		if err != nil && err == io.EOF {
			if err == io.EOF {
				//client close socket
				s.Close()
				return
			}

			//TODO
			panic(err)
		}

		s.netOp.PostEvent(&iface.MessageEvent{EventSes: s, Msg: msg})
	}
}

func (s *TcpSession) writeLoop() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("catch unexpected error:%s", r)
			s.Close()
		}

		s.wg.Done()
	}()

	var err error
	for msg := range s.sendCh {
		if msg == nil {
			break
		}

		err = s.netOp.WriteMessage(s.raw, msg)
		if err != nil {
			//TODO
			panic(err)
		}
	}
}

func (s *TcpSession) LoadCtx(k interface{}) (v interface{}, ok bool) {
	return s.ctx.Load(k)
}

func (s *TcpSession) StoreCtx(k, v interface{}) {
	s.ctx.Store(k, v)
}

func (s *TcpSession) Send(msg iface.Message) {
	//TODO make it never blocking
	s.sendCh <- msg
}
