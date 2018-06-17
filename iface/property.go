package iface

type Property interface {
	LoadCtx(key interface{}) (val interface{}, ok bool)
	StoreCtx(key interface{}, val interface{})
}
