package message_meta

import "reflect"

type MessageMeta struct {
	ID   uint32
	Type reflect.Type
}

func NewMessageMeta(id uint32, pType reflect.Type) *MessageMeta {
	return &MessageMeta{
		ID:   id,
		Type: pType,
	}
}

func (m *MessageMeta) NewType() interface{} {
	if m.Type.Kind() == reflect.Ptr {
		m.Type = m.Type.Elem()
	}

	return reflect.New(m.Type).Interface()
}
