package message

import "reflect"

type MessageMeta struct {
	ID   uint32
	Type reflect.Type
}

func (m *MessageMeta) NewType() interface{} {

	if m.Type.Kind() == reflect.Ptr {
		m.Type = m.Type.Elem()
	}

	return reflect.New(m.Type).Interface()
}
