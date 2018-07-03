package message_meta

import "fmt"

var (
	metas = map[uint32]*MessageMeta{}
)

// RegisterMsgMeta 注册一个meta
// 如果meta已存在或meda id重复,则panic
func RegisterMsgMeta(m *MessageMeta) {
	if _, ok := metas[m.ID]; ok {
		panic(fmt.Sprintf("dup register message_meta meta, id :%d", m.ID))
	}

	metas[m.ID] = m
}

// MustGetMsgMeta 获取指定id对应的meta.
// 若未注册,则panic
func MustGetMsgMeta(id uint32) *MessageMeta {

	if m, ok := metas[id]; ok {
		return m
	}

	panic(fmt.Sprintf("message_meta meta not register , id :%d", id))
}
