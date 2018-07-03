package worker_pool

import "github.com/MaxnSter/gnet/iface"

// Pool 是一个goroutine pool,
// 单独使用可用于限制并发数量,
// 作为gnet组件时还可以指定并发模型
type Pool interface {
	// Start启动pool,此方法保证goroutineeeee safe
	Start()

	// Stop停止pool,调用方阻塞直到Stop返回
	// pool保证此时剩余的pool item全部执行完毕才返回
	Stop()

	// StopAsync与Stop相同,但它立即返回, pool完全停止时done active
	StopAsync() (done <-chan struct{})

	// Put往pool中投放任务,无论pool是否已满,此次投放必定成功
	Put(ctx iface.Context, cb func(iface.Context))

	// TryPut与Put相同,但当pool已满试,投放失败,返回false
	TryPut(ctx iface.Context, cb func(ctx iface.Context)) bool

	// TypeName返回pool的唯一表示
	TypeName() string
}
