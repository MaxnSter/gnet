package meta

import "reflect"

// meta works well with type_length_value plugins
type Meta interface {
	Identify() uint32
	Type() reflect.Type
	New() interface{}
}

func New(id uint32, t reflect.Type) Meta {
	return metaWrapper{
		id:    id,
		pType: t,
	}
}

type metaWrapper struct {
	id    uint32
	pType reflect.Type
}

func (m metaWrapper) Identify() uint32 {
	return m.id
}

func (m metaWrapper) Type() reflect.Type {
	return m.pType
}

func (m metaWrapper) New() interface{} {
	t := m.pType
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return reflect.New(t).Interface()
}
