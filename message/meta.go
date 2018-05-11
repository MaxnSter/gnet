package message

import "reflect"

type messageMeta struct {
	ID   uint32
	Type reflect.Type
}

func NewMsgMeta(id uint32, pType reflect.Type) *messageMeta {
	return &messageMeta{
		ID:   id,
		Type: pType,
	}
}

func (m *messageMeta) NewType() interface{} {

	if m.Type.Kind() == reflect.Ptr {
		m.Type = m.Type.Elem()
	}

	return reflect.New(m.Type).Interface()
}
