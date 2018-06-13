package tcp

import (
	"bufio"
	"io"
	"net"
	"runtime/debug"
	"sync"
	"time"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/gnet/timer"
	"github.com/MaxnSter/gnet/util"
	"github.com/sirupsen/logrus"
)

const (
	//TODO high water mark
	//发送缓冲区
	sendBuf = 128

	//shutdown write之后,给readLoop一个超时,指定时间内未触发io.EOF则强制关闭
	rdTimeout = time.Second * 5
)

var (
	_ gnet.NetSession = (*tcpSession)(nil)
)

type tcpSession struct {
	id       int64
	ctx      sync.Map
	raw      *net.TCPConn
	connWrap *bufio.ReadWriter
	sendQue  *util.MsgQueue
	status   *status

	wg          *sync.WaitGroup
	closeCh     chan struct{}
	onCloseDone func(*tcpSession)

	module   gnet.Module
	operator gnet.Operator
}

func NewTcpSession(id int64, conn *net.TCPConn, m gnet.Module, o gnet.Operator, onCloseDone func(*tcpSession)) *tcpSession {
	s := &tcpSession{
		id:          id,
		raw:         conn,
		onCloseDone: onCloseDone,
		closeCh:     make(chan struct{}),
		module:      m,
		operator:    o,
		sendQue:     util.NewMsgQueueWithCap(sendBuf),
		wg:          &sync.WaitGroup{},
		status:      &status{},
	}

	s.connWrap = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	return s
}

func (s *tcpSession) Raw() io.ReadWriter {
	return s.raw
}

func (s *tcpSession) ID() int64 {
	return s.id
}

func (s *tcpSession) Start() {
	if !s.status.start() {
		return
	}
	logger.WithField("sessionId", s.id).Debugln("session start")

	//start loop
	s.wg.Add(2)
	go s.readLoop()
	go s.writeLoop()

	//callback to user
	if s.operator.GetOnConnected() != nil {
		logger.WithField("sessionId", s.id).Debugln("session onConnected, callback to user")
		s.operator.GetOnConnected()(s)
	}

}

// 正确关闭tcp连接的做法
// correct  sender:		send() + shutdown(wr) + read()->0 + close socket
// correct  receiver:	read()->0 + nothing more to send -> close socket
// 流程如下->
// sender:shutdown(wr) -> receiver:read(0) -> receiver:send over and close socket ->
// sender:read(0) -> sender:close socket -> socket正常关闭
func (s *tcpSession) Stop() {
	if !s.status.stop() {
		return
	}

	go func() {

		logger.WithField("sessionId", s.id).Debugln("session stopping...")

		if s.operator.GetOnClose() != nil {
			logger.WithField("sessionId", s.id).Debugln("session onClose, callback to user")
			s.operator.GetOnClose()(s)
		}

		//close readLoop
		logger.WithField("sessionId", s.id).Debugln("close readLoop...")
		close(s.closeCh)

		//send signal to writeLoop, shutdown wr until nothing more to send
		logger.WithField("sessionId", s.id).Debugln("close WriteLoop...")
		s.sendQue.Put(nil)

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

func (s *tcpSession) readLoop() {
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

		msg, err := s.operator.Read(s.connWrap, s.module)
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

		s.operator.PostEvent(&gnet.EventWrapper{EventSession: s, Msg: msg}, s.module)
	}
}

func (s *tcpSession) writeLoop() {
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
				s.connWrap.Flush()
				return
			}

			err = s.operator.Write(s.connWrap, msg, s.module)
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

func (s *tcpSession) Send(msg interface{}) {
	//select {
	//case <-s.closeCh:
	//	return
	//default:
	//}
	s.sendQue.Put(msg)
}

func (s *tcpSession) LoadCtx(k interface{}) (v interface{}, ok bool) {
	return s.ctx.Load(k)
}

func (s *tcpSession) StoreCtx(k, v interface{}) {
	s.ctx.Store(k, v)
}

func (s *tcpSession) RunAt(runAt time.Time, cb timer.OnTimeOut) (timerId int64) {
	if s.module.Timer() == nil {
		return -1
	}

	return s.module.Timer().AddTimer(runAt, 0, s, cb)
}

func (s *tcpSession) RunAfter(after time.Duration, cb timer.OnTimeOut) (timerId int64) {
	if s.module.Timer() == nil {
		return -1
	}
	return s.module.Timer().AddTimer(time.Now().Add(after), 0, s, cb)
}

func (s *tcpSession) RunEvery(runAt time.Time, interval time.Duration, cb timer.OnTimeOut) (timerId int64) {
	if s.module.Timer() == nil {
		return -1
	}
	return s.module.Timer().AddTimer(runAt, interval, s, cb)
}

func (s *tcpSession) CancelTimer(timerId int64) {
	if s.module.Timer() == nil {
		return
	}

	s.module.Timer().CancelTimer(timerId)
}
