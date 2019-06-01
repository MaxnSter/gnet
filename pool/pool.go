package pool

// Pool 是一个goroutine pool,
// 单独使用可用于限制并发数量,
// 作为gnet组件时还可以指定并发模型
type Pool interface {
	Run()
	// Stop停止pool,调用方阻塞直到Stop返回
	// pool保证此时剩余的pool item全部执行完毕才返回
	Stop()

	// Put往pool中投放任务,无论pool是否已满,此次投放必定成功
	Put(f func(), opts ...func(*Option))
	// TryPut与Put相同,但当pool已满试,投放失败,返回false
	TryPut(f func(), opts ...func(*Option)) bool

	String() string
}

type Identifier interface {
	ID() uint64
}

type Option struct {
	Identifier Identifier
}

func WithIdentify(i Identifier) func(*Option) {
	return func(option *Option) {
		option.Identifier = i
	}
}
