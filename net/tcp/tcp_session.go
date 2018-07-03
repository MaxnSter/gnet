package tcp

import (
	"bufio"
	"io"
	"net"
	"runtime/debug"
	"sync"
	"time"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/iface"
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
	ctx      sync.Map //FIXME 过于粗暴
	raw      *net.TCPConn
	connWrap *bufio.ReadWriter //save syscall
	sendQue  *util.MsgQueue
	status   *status

	wg          *sync.WaitGroup
	closeCh     chan struct{}
	onCloseDone func(*tcpSession)

	manager  gnet.SessionManager
	module   gnet.Module
	operator gnet.Operator
}

func newTCPSession(id int64, conn *net.TCPConn, mg gnet.SessionManager, m gnet.Module, o gnet.Operator, onCloseDone func(*tcpSession)) *tcpSession {
	s := &tcpSession{
		id:          id,
		raw:         conn,
		onCloseDone: onCloseDone,
		closeCh:     make(chan struct{}),
		manager:     mg,
		module:      m,
		operator:    o,
		sendQue:     util.NewMsgQueueWithCap(sendBuf),
		wg:          &sync.WaitGroup{},
		status:      &status{},
	}

	s.connWrap = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	return s
}

// AccessManager返回管理当前NetSession的SessionManager
func (s *tcpSession) AccessManager() gnet.SessionManager {
	return s.manager
}

// Raw返回当前NetSession对应的读写接口
// 为了降低调用者的使用权限,有意不返回net.conn,不过当然可以type assert...
func (s *tcpSession) Raw() io.ReadWriter {
	return s.raw
}

// ID返回当前NetSession对应的ID标识
func (s *tcpSession) ID() int64 {
	return s.id
}

func (s *tcpSession) start() {
	if !s.status.start() {
		return
	}
	logger.WithField("sessionId", s.id).Debugln("session start")

	//start loop
	s.wg.Add(2)
	go s.readLoop()
	go s.writeLoop()

	//callback to user in loop
	if s.operator.GetOnConnected() != nil {
		if s.module.Pool() == nil {
			s.operator.GetOnConnected()(s)
		} else {
			s.RunInPool(s.operator.GetOnConnected())
		}
	}

}

// 关于正确关闭tcp连接的看法
// 发送方:		send() + shutdown(wr) + read()->0 + close socket
// 接收方:		read()->0 + nothing more to send -> close socket
// 流程如下->
// sender:shutdown(wr) -> receiver:read(0) -> receiver:send over and close socket ->
// sender:read(0) -> sender:close socket -> socket正常关闭

// Stop关闭当前连接,此调用立即返回,不会等待连接关闭完成
func (s *tcpSession) Stop() {
	if !s.status.stop() {
		return
	}

	go func() {

		logger.WithField("sessionId", s.id).Debugln("session stopping...")

		//用户回调
		if s.operator.GetOnClose() != nil {
			logger.WithField("sessionId", s.id).Debugln("session onClose, callback to user")
			if s.module.Pool() == nil {
				s.operator.GetOnClose()(s)
			} else {
				s.RunInPool(s.operator.GetOnClose())
			}
		}

		//关闭读端
		logger.WithField("sessionId", s.id).Debugln("close readLoop...")
		close(s.closeCh)

		//关闭写端
		logger.WithField("sessionId", s.id).Debugln("close WriteLoop...")
		s.sendQue.Put(nil)

		//等待读写两端关闭完成
		s.wg.Wait()

		//关闭文件描述符
		logger.WithField("sessionId", s.id).Debugln("session closeDone, close raw socket")
		s.raw.Close()

		//通知sessionManager
		if s.onCloseDone != nil {
			logger.WithField("sessionId", s.id).Debugln("session closeDone, notify server")
			s.onCloseDone(s)
		}
	}()
}

func (s *tcpSession) readLoop() {
	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			s.Stop()
		}

		logger.WithField("sessionId", s.id).Debugln("readLoop stopped")
		s.wg.Done()
	}()

	logger.WithField("sessionId", s.id).Debugln("session start readLoop")

	for {
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

			//未知错误
			logger.WithFields(logrus.Fields{"sessionId": s.id, "error": err}).Errorln("session readLoop error")
			s.Stop()
			return
		}

		s.operator.PostEvent(&gnet.EventWrapper{EventSession: s, Msg: msg}, s.module)
	}
}

func (s *tcpSession) writeLoop() {
	defer func() {
		if r := recover(); r != nil {
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
				logger.WithFields(logrus.Fields{"sessionId": s.id, "error": err}).
					Errorln("session writeLoop error")
				s.Stop()
				return
			}
		}

		err = s.connWrap.Flush()
		if err != nil {
			logger.WithFields(logrus.Fields{"sessionId": s.id, "error": err}).
				Errorln("session writeLoop error")
			s.Stop()
			return
		}
	}
}

// RunInPool将f投入module对应的工作池中异步执行
// 若module未设置pool,则直接执行f
func (s *tcpSession) RunInPool(f func(gnet.NetSession)) {
	if s.module.Pool() == nil {
		f(s)
		return
	}
	s.module.Pool().Put(s, func(ctx iface.Context) {
		f(ctx.(gnet.NetSession))
	})
}

// Send添加消息至发送队列,保证goroutine safe,不阻塞
func (s *tcpSession) Send(msg interface{}) {
	s.sendQue.Put(msg)
}

// LoadCtx加载key对应的上下文
func (s *tcpSession) LoadCtx(k interface{}) (v interface{}, ok bool) {
	return s.ctx.Load(k)
}

// StoreCtx保存上下文信息
func (s *tcpSession) StoreCtx(k, v interface{}) {
	s.ctx.Store(k, v)
}

// RunAt添加一个单次定时器,在runAt时间触发cb
// 注意:cb中的ctx为NetSession
// 若module未指定timer,则此调用无效
func (s *tcpSession) RunAt(runAt time.Time, cb timer.OnTimeOut) (timerID int64) {
	if s.module.Timer() == nil {
		return -1
	}

	return s.module.Timer().AddTimer(runAt, 0, s, cb)
}

// RunAfter添加一个单次定时器,在Now + After时间触发cb
// 注意:cb中的ctx为NetSession
// 若module未指定timer,则此调用无效
func (s *tcpSession) RunAfter(after time.Duration, cb timer.OnTimeOut) (timerID int64) {
	if s.module.Timer() == nil {
		return -1
	}
	return s.module.Timer().AddTimer(time.Now().Add(after), 0, s, cb)
}

// RunEvery增加一个interval执行周期的定时器,在runAt触发第一次cb
// 注意:cb中的ctx为NetSession
// 若module未指定timer,则此调用无效
func (s *tcpSession) RunEvery(runAt time.Time, interval time.Duration, cb timer.OnTimeOut) (timerID int64) {
	if s.module.Timer() == nil {
		return -1
	}
	return s.module.Timer().AddTimer(runAt, interval, s, cb)
}

// CancelTimer取消timerId对应的定时器
// 若定时器已触发或timerId无效,则次调用无效
func (s *tcpSession) CancelTimer(timerId int64) {
	if s.module.Timer() == nil {
		return
	}

	s.module.Timer().CancelTimer(timerId)
}
