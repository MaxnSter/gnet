package gnet

import (
	"io"
	"time"

	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/timer"
)

type SessionManager interface {
	// BroadCast对所有NetSession连接执行fn
	// 若module设置Pool,则fn全部投入Pool中,否则在当前goroutine执行
	Broadcast(func(session NetSession))

	// GetSession返回指定id对应的NetSession
	GetSession(id int64) (NetSession, bool)
}

// ModuleRunner 表示一个module持有者,提供module中的pool,timer的使用接口
type ModuleRunner interface {
	// RunInPool将f投入module对应的工作池中异步执行
	// 若module未设置pool,则直接执行f
	RunInPool(func(NetSession))

	// RunAt添加一个单次定时器,在runAt时间触发cb
	// 注意:cb中的ctx为NetSession
	// 若module未指定timer,则此调用无效
	//TODO 若module未指定timer,则使用标准库的timer
	RunAt(at time.Time, cb timer.OnTimeOut) (timerID int64)

	// RunAfter添加一个单次定时器,在Now + After时间触发cb
	// 注意:cb中的ctx为NetSession
	// 若module未指定timer,则此调用无效
	RunAfter(after time.Duration, cb timer.OnTimeOut) (timerID int64)

	// RunEvery增加一个interval执行周期的定时器,在runAt触发第一次cb
	// 注意:cb中的ctx为NetSession
	// 若module未指定timer,则此调用无效
	RunEvery(at time.Time, interval time.Duration, cb timer.OnTimeOut) (timerID int64)

	// CancelTimer取消timerId对应的定时器
	// 若定时器已触发或timerId无效,则次调用无效
	CancelTimer(timerID int64)
}

type NetSession interface {
	iface.Identifier
	iface.Property
	ModuleRunner

	// Raw返回当前NetSession对应的读写接口
	// 为了降低调用者的使用权限,有意不返回net.conn,不过当然可以type assert...
	Raw() io.ReadWriter
	// Send添加消息至发送队列,保证goroutine safe,不阻塞
	Send(message interface{})
	// AccessManager返回管理当前NetSession的SessionManager
	AccessManager() SessionManager
	// Stop关闭当前连接,此调用立即返回,不会等待连接关闭完成
	Stop()
}

type NetServer interface {
	SessionManager

	// Serve启动服务器,调用方阻塞直到服务器关闭完成
	// Serve必须在Listen成功后才可调用
	Serve()

	// Listen开始监听指定的addr
	Listen(addr string) error

	// ListenAndServe监听指定addr并启动服务器
	// 若Listen成功,则阻塞直到服务器关闭完成
	// 若Listen失败,则panic
	ListenAndServe(addr string)

	// Stop停止服务器,调用方阻塞知道服务器所有模块关闭完成
	Stop()
}

type NetClient interface {
	SessionManager
	// SetSessionNumber设置客户端连接数,默认为1
	SetSessionNumber(sessionNumber int)

	// Connect开始建立连接并启动客户端,本次调用将阻塞直到客户端退出
	// Connect会自动重试直到成功建立连接,有意不指定最大重连次数
	Connect(addr string)

	// Stop停止客户端,关闭当前所有客户端连接
	Stop()
}
