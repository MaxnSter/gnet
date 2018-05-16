package net

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/MaxnSter/gnet/iface"
)

const (
	//TODO high water mark and small water mark
	tcpSendBuf = 100
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
	raw     *net.TCPConn
	ctx     sync.Map

	onCloseDone func(*TcpSession)
}

func NewTcpSession(id int64, netOp *NetOptions, conn *net.TCPConn, onCloseDone func(*TcpSession)) *TcpSession {
	return &TcpSession{
		id:          id,
		netOp:       netOp,
		raw:         conn,
		onCloseDone: onCloseDone,
		closeCh:     make(chan struct{}),
		sendCh:      make(chan iface.Message, tcpSendBuf),
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
	if s.netOp.OnConnected != nil {
		s.netOp.OnConnected(s)
	}

	//wait for readLoop and writeLoop finish
	s.wg.Wait()

	//close socket
	s.raw.Close()

	//tell sessionManager we are done
	if s.onCloseDone != nil {
		s.onCloseDone(s)
	}

}

// 正确关闭tcp连接的做法
// correct  sender:		send() + shutdown(wr) + read()->0 + close socket
// correct  receiver:	read()->0 + nothing more to send -> close socket
// 流程如下->
// sender:shutdown(wr) -> receiver:read(0) -> receiver:send over and close socket ->
// sender:read(0) -> sender:close socket -> socket正常关闭
func (s *TcpSession) Stop() {
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

	//send signal to writeLoop, shutdown wr until nothing more to send
	s.sendCh <- nil
}

func (s *TcpSession) readLoop() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("catch unexpected error:%s", r)
			s.Stop()
		}

		s.wg.Done()
	}()

	for {
		select {
		case <-s.closeCh:
		default:
		}

		msg, err := s.netOp.ReadMessage(s.raw)
		if err != nil {

			if err == io.EOF {
				//client close socket
				s.Stop()
				return
			}

			//TODO logs
			panic(err)
		}

		s.netOp.PostEvent(&iface.MessageEvent{EventSes: s, Msg: msg})
	}
}

func (s *TcpSession) writeLoop() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("catch unexpected error:%s", r)

			//通知readLoop 关闭
			s.Stop()
		}

		//shutdown write
		//TODO shutdown write之后,给readLoop一个超时,指定时间内未触发io.EOF则强制关闭
		s.raw.CloseWrite()
		s.wg.Done()
	}()

	var err error
	for msg := range s.sendCh {
		if msg == nil {
			break
		}

		err = s.netOp.WriteMessage(s.raw, msg)
		if err != nil {
			//TODO logs
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
	s.sendCh <- msg
}

func (s *TcpSession) RunAt(runAt time.Time, cb iface.TimeOutCB) (timerId int64) {
	return s.netOp.Timer.AddTimer(runAt, 0, s, cb)
}

func (s *TcpSession) RunAfter(start time.Time, after time.Duration, cb iface.TimeOutCB) (timerId int64) {
	return s.netOp.Timer.AddTimer(start.Add(after), 0, s, cb)
}

func (s *TcpSession) RunEvery(runAt time.Time, interval time.Duration, cb iface.TimeOutCB) (timerId int64) {
	return s.netOp.Timer.AddTimer(runAt, interval, s, cb)
}

func (s *TcpSession) CancelTimer(timerId int64) {
	s.netOp.Timer.CancelTimer(timerId)
}
