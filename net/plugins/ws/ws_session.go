package ws

import (
	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/gnet/timer"
	"github.com/MaxnSter/gnet/util"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"runtime/debug"
	"sync"
	"time"
)

type wsSession struct {
	module   gnet.Module
	operator gnet.Operator

	manager     gnet.SessionManager
	onCloseDone func(*wsSession)
	onClose     chan struct{}

	conn *websocket.Conn

	id    int64
	wg    sync.WaitGroup
	sendQ *util.MsgQueue
	ctx   map[interface{}]interface{}
}

func newWsSession(id int64, m gnet.Module, o gnet.Operator,
	mg gnet.SessionManager, w http.ResponseWriter, r *http.Request,
	onDone func(*wsSession)) gnet.NetSession {
	ws := &wsSession{
		id:          id,
		module:      m,
		operator:    o,
		sendQ:       util.NewMsgQueue(),
		wg:          sync.WaitGroup{},
		manager:     mg,
		onCloseDone: onDone,
		onClose:     make(chan struct{}),
		ctx:         make(map[interface{}]interface{}),
	}

	var err error
	ws.conn, err = upgrader.Upgrade(w, r, nil)

	if err != nil {
		ws.onCloseDone(ws)
	} else {
		go ws.run()
	}

	return ws
}

func (ws *wsSession) run() {
	ws.wg.Add(2)

	go ws.readLoop()
	go ws.writeLoop()

	if ws.operator.GetOnConnected() != nil {
		if ws.module.Pool() == nil {
			ws.operator.GetOnConnected()(ws)
		} else {
			logger.Debugln("call back to user")
			ws.RunInPool(ws.operator.GetOnConnected())
		}
	}

	ws.wg.Wait()
	ws.conn.Close()
	ws.onCloseDone(ws)
}

func (ws *wsSession) readLoop() {
	defer func() {
		if r := recover(); r != nil {
			// FIXME
			logger.Warning(r)
		}

		ws.wg.Done()
	}()

	for {
		_, reader, err := ws.conn.NextReader()
		if err != nil {
			//FIXME log
			logger.Debug(err)
			ws.Stop()
			return
		}

		//FIXME
		msg, err := ws.operator.Read(reader, ws.module)
		if err != nil {
			//FIXME log
			logger.Debug(err)
			debug.PrintStack()
			ws.Stop()
			return
		}

		logger.Debugln("recv:" , msg)
		ws.operator.PostEvent(&gnet.EventWrapper{ws, msg}, ws.module)
	}
}

func (ws *wsSession) writeLoop() {
	defer func() {
		if r := recover(); r != nil {
			// FIXME
			logger.Warning(r)
		}

		//主动断开连接后防止写超时
		ws.conn.SetReadDeadline(time.Now().Add(3 * time.Second))

		ws.wg.Done()
	}()

	var msgs []interface{}
	for {
		select {
		case <-ws.onClose:
			return
		default:
		}

		wr, err := ws.conn.NextWriter(websocket.BinaryMessage)
		if err != nil {
			logger.Debug(err)
			ws.Stop()
			return
		}

		ws.sendQ.PickWithSignal(ws.onClose, &msgs)

		for _, msg := range msgs {
			if msg == nil {
				logger.Debug("session stop")
				return
			}

			logger.Debugln("send:", msg)
			err := ws.operator.Write(wr, msg, ws.module)
			if err != nil {
				logger.Debug(err)
				ws.Stop()
				return
			}
		}

		msgs = msgs[0:0]
	}
}

// Raw返回当前NetSession对应的读写接口
// 为了降低调用者的使用权限,有意不返回net.conn,不过当然可以type assert...
func (ws *wsSession) Addr() net.Addr {
	return ws.conn.RemoteAddr()
}

// Send添加消息至发送队列,保证goroutine safe,不阻塞
func (ws *wsSession) Send(message interface{}) {
	ws.sendQ.Put(message)
}

// AccessManager返回管理当前NetSession的SessionManager
func (ws *wsSession) AccessManager() gnet.SessionManager {
	return ws.manager
}

// Stop关闭当前连接,此调用立即返回,不会等待连接关闭完成
func (ws *wsSession) Stop() {
	select {
	case <-ws.onClose:
	default:
		close(ws.onClose)
	}
}

func (ws *wsSession) ID() int64 {
	return ws.id
}

func (ws *wsSession) LoadCtx(key interface{}) (val interface{}, ok bool) {
	// FIXME goroutine safe
	val, ok = ws.ctx[key]
	return
}

// StoreCtx存放一个上下文对象
func (ws *wsSession) StoreCtx(key interface{}, val interface{}) {
	// FIXME goroutine safe
	ws.ctx[key] = val
}

// RunInPool将f投入module对应的工作池中异步执行
// 若module未设置pool,则直接执行f
func (ws *wsSession) RunInPool(f func(gnet.NetSession)) {
	if ws.module.Pool() == nil {
		f(ws)
		return
	}

	ws.module.Pool().Put(nil, func(context iface.Context) {
		f(ws)
	})
}

// RunAt添加一个单次定时器,在runAt时间触发cb
// 注意:cb中的ctx为NetSession
// 若module未指定timer,则此调用无效
//TODO 若module未指定timer,则使用标准库的timer
func (ws *wsSession) RunAt(at time.Time, cb timer.OnTimeOut) (timerID int64) {
	return ws.module.Timer().AddTimer(at, 0, ws, cb)
}

// RunAfter添加一个单次定时器,在Now + After时间触发cb
// 注意:cb中的ctx为NetSession
// 若module未指定timer,则此调用无效
func (ws *wsSession) RunAfter(after time.Duration, cb timer.OnTimeOut) (timerID int64) {
	return ws.RunAt(time.Now().Add(after), cb)
}

// RunEvery增加一个interval执行周期的定时器,在runAt触发第一次cb
// 注意:cb中的ctx为NetSession
// 若module未指定timer,则此调用无效
func (ws *wsSession) RunEvery(at time.Time, interval time.Duration, cb timer.OnTimeOut) (timerID int64) {
	return ws.module.Timer().AddTimer(at, interval, ws, cb)
}

// CancelTimer取消timerId对应的定时器
// 若定时器已触发或timerId无效,则次调用无效
func (ws *wsSession) CancelTimer(timerID int64) {
	ws.module.Timer().CancelTimer(timerID)
}
