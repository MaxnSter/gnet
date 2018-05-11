package message

import "fmt"

var (
	metas = map[uint32]*messageMeta{}
)

func RegisterMsgMeta(id uint32, m *messageMeta) {
	if _, ok := metas[id]; ok {
		panic(fmt.Sprintf("dup register message meta, id :%d", id))
	}

	metas[id] = m
}

func GetMsgMeta(id uint32) *messageMeta {
	if m, ok := metas[id]; ok {
		return m
	}

	return nil
}

func MustGetMsgMeta(id uint32) *messageMeta {

	if m, ok := metas[id]; ok {
		return m
	}

	panic(fmt.Sprintf("message meta not register , id :%d", id))
}
