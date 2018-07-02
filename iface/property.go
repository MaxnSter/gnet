package iface

// Property表示一个可以设置上下文的对象
type Property interface {
	// LoadCtx获取一个指定的上下文对象
	LoadCtx(key interface{}) (val interface{}, ok bool)

	// StoreCtx存放一个上下文对象
	StoreCtx(key interface{}, val interface{})
}
