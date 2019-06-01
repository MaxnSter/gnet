package meta

import "fmt"

var (
	metas = map[uint32]Meta{}
)

// RegisterMsgMeta 注册一个meta
// 如果meta已存在或meda id重复,则panic
func RegisterMsgMeta(m Meta) {
	if _, ok := metas[m.Identify()]; ok {
		panic(fmt.Sprintf("dup register message_meta meta, id :%d", m.Identify()))
	}

	metas[m.Identify()] = m
}

// MustGetMsgMeta 获取指定id对应的meta.
// 若未注册,则panic
func MustGetMsgMeta(id uint32) Meta {

	if m, ok := metas[id]; ok {
		return m
	}

	panic(fmt.Sprintf("message_meta meta not register , id :%d", id))
}
