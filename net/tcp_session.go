package net

import (
	"bufio"
	"io"
	"net"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/gnet/util"
	"github.com/sirupsen/logrus"
)

const (
	//TODO high water mark and small water mark
	//发送缓冲区
	sendBuf = 128

	//shutdown write之后,给readLoop一个超时,指定时间内未触发io.EOF则强制关闭
	rdTimeout = time.Second * 5
)

type sessionStatus = int64

const (
	ready sessionStatus = iota
	start
	stop
)

var (
	_ iface.NetSession = (*TcpSession)(nil)
)

type TcpSession struct {
	id      int64
	manager *TcpServer

	netOp *NetOptions
	ctx   sync.Map

	status  sessionStatus
	guard   *sync.Mutex
	started bool
	closed  bool

	raw      *net.TCPConn
	connWrap *bufio.ReadWriter
	sendQue  *util.MsgQueue

	wg          *sync.WaitGroup
	closeCh     chan struct{}
	onCloseDone func(*TcpSession)
}

func NewTcpSession(id int64, netOp *NetOptions, conn *net.TCPConn, onCloseDone func(*TcpSession)) *TcpSession {
	s := &TcpSession{
		id:          id,
		netOp:       netOp,
		raw:         conn,
		onCloseDone: onCloseDone,
		closeCh:     make(chan struct{}),
		sendQue:     util.NewMsgQueueWithCap(sendBuf),
		wg:          &sync.WaitGroup{},
		guard:       &sync.Mutex{},
	}

	s.connWrap = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	return s
}

func (s *TcpSession) SetManager(m *TcpServer) {
	s.manager = m
}

func (s *TcpSession) ID() int64 {
	return s.id
}

func (s *TcpSession) Start() {
	if !atomic.CompareAndSwapInt64(&s.status, ready, start) {
		return
	}
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

}

// 正确关闭tcp连接的做法
// correct  sender:		send() + shutdown(wr) + read()->0 + close socket
// correct  receiver:	read()->0 + nothing more to send -> close socket
// 流程如下->
// sender:shutdown(wr) -> receiver:read(0) -> receiver:send over and close socket ->
// sender:read(0) -> sender:close socket -> socket正常关闭
func (s *TcpSession) Stop() {
	if !atomic.CompareAndSwapInt64(&s.status, start, stop) {
		return
	}

	go func() {

		logger.WithField("sessionId", s.id).Debugln("session stopping...")

		if s.netOp.OnSessionClose != nil {
			logger.WithField("sessionId", s.id).Debugln("session onClose, callback to user")
			s.netOp.OnSessionClose(s)
		}

		//close readLoop
		logger.WithField("sessionId", s.id).Debugln("close readLoop...")
		close(s.closeCh)

		//send signal to writeLoop, shutdown wr until nothing more to send
		logger.WithField("sessionId", s.id).Debugln("close WriteLoop...")
		s.sendQue.Add(nil)

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
	}()
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

		msg, err := s.netOp.ReadMessage(s.connWrap)
		if err != nil {

			//remote close socket
			if err == io.EOF {
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
			debug.PrintStack()
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
	var msgs []interface{}

	for {
		msgs = msgs[0:0]
		s.sendQue.PickWithSignal(s.closeCh, &msgs)

		for _, msg := range msgs {

			if msg == nil {
				//TODO error handing?
				s.connWrap.Flush()
				return
			}

			err = s.netOp.WriteMessage(s.connWrap, msg)
			if err != nil {
				panic(err)
			}

		}

		err = s.connWrap.Flush()
		if err != nil {
			panic(err)
		}
	}

}

func (s *TcpSession) Send(msg interface{}) {

	select {
	case <-s.closeCh:
		return
	default:
	}

	s.sendQue.Add(msg)
}

func (s *TcpSession) LoadCtx(k interface{}) (v interface{}, ok bool) {
	return s.ctx.Load(k)
}

func (s *TcpSession) StoreCtx(k, v interface{}) {
	s.ctx.Store(k, v)
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
