package message_meta

import "fmt"

var (
	metas = map[uint32]*MessageMeta{}
)


func RegisterMsgMeta(m *MessageMeta) {
	if _, ok := metas[m.ID]; ok {
		panic(fmt.Sprintf("dup register message_meta meta, id :%d", m.ID))
	}

	metas[m.ID] = m
}

func MustGetMsgMeta(id uint32) *MessageMeta {

	if m, ok := metas[id]; ok {
		return m
	}

	panic(fmt.Sprintf("message_meta meta not register , id :%d", id))
}
