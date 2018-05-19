package net

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/logger"
	"github.com/sirupsen/logrus"
)

const (
	//TODO high water mark and small water mark
	//发送缓冲区
	sendBuf = 100

	//shutdown write之后,给readLoop一个超时,指定时间内未触发io.EOF则强制关闭
	rdTimeout = time.Second * 5
)

var (
	_ iface.NetSession = (*TcpSession)(nil)
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
		sendCh:      make(chan iface.Message, sendBuf),
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

	logger.WithField("sessionId", s.id).Debugln("session start")

	//start loop
	s.wg.Add(2)
	go s.readLoop()
	go s.writeLoop()

	//callback to user
	if s.netOp.OnConnected != nil {
		logger.WithField("sessionId", s.id).Debugln("session onConnected, callback to user")
		s.netOp.OnConnected(s)
	}

	//wait for readLoop and writeLoop finish
	s.wg.Wait()

	//close socket
	logger.WithField("sessionId", s.id).Debugln("session closeDone, close raw socket")
	s.raw.Close()

	//tell sessionManager we are done
	if s.onCloseDone != nil {
		logger.WithField("sessionId", s.id).Debugln("session closeDone, notify server")
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

	logger.WithField("sessionId", s.id).Debugln("session stopping...")

	if s.netOp.OnClose != nil {
		logger.WithField("sessionId", s.id).Debugln("session onClose, callback to user")
		s.netOp.OnClose(s)
	}

	//close readLoop
	logger.WithField("sessionId", s.id).Debugln("close readLoop...")
	close(s.closeCh)

	//send signal to writeLoop, shutdown wr until nothing more to send
	logger.WithField("sessionId", s.id).Debugln("close WriteLoop...")
	s.sendCh <- nil
}

func (s *TcpSession) readLoop() {
	defer func() {
		if r := recover(); r != nil {
			logger.WithFields(logrus.Fields{
				"sessionId": s.id,
				"error":     r}).Errorln("session readLoop error")

			s.Stop()
		}

		logger.WithField("sessionId", s.id).Debugln("readLoop stopped")
		s.wg.Done()
	}()

	logger.WithField("sessionId", s.id).Debugln("session start readLoop")

	for {
		select {
		case <-s.closeCh:
			return
		default:
		}

		msg, err := s.netOp.ReadMessage(s.raw)
		if err != nil {

			if err == io.EOF {
				//client close socket
				logger.WithField("sessionId", s.id).Debugln("read eof from socket")
				s.Stop()
				return
			}

			if err, ok := err.(net.Error); ok && err.Timeout() {
				select {
				case <-s.closeCh:
					//session已经调用过stop, 证明这里的timeout是writeLoop发起的,正常退出
					return
				default:
				}
			}

			panic(err)
		}

		logger.WithFields(logrus.Fields{
			"sessionId": s.id,
			"messageId": msg.GetId(),
		}).Debugln("receive message from socket")

		s.netOp.PostEvent(&iface.MessageEvent{EventSes: s, Msg: msg})
	}
}

func (s *TcpSession) writeLoop() {
	defer func() {
		if r := recover(); r != nil {
			logger.WithFields(logrus.Fields{
				"sessionId": s.id,
				"error":     r,
			}).Errorln("session writeLoop error")

			//通知readLoop 关闭
			s.Stop()
		}

		//shutdown write
		s.raw.CloseWrite()
		s.raw.SetReadDeadline(time.Now().Add(rdTimeout))
		s.wg.Done()

		logger.WithField("sessionId", s.id).Debugln("writeLoop stopped")
	}()

	logger.WithField("sessionId", s.id).Debugln("session start writeLoop")

	var err error
	for msg := range s.sendCh {
		if msg == nil {
			break
		}

		err = s.netOp.WriteMessage(s.raw, msg)
		if err != nil {
			panic(err)
		}

		logger.WithFields(logrus.Fields{
			"sessionId": s.id,
			"messageId": msg.GetId(),
		}).Debugln("send message to socket")
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
