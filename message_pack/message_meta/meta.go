package message_meta

import "reflect"

// MessageMeta 为消息元对象
// 主要和tlv Packer配合使用,让业务方无需关心消息的decode
// 次模块参考自cellnet
type MessageMeta struct {
	// ID为该消息元的唯一标志 每个消息元必须唯一
	ID uint32 //TODO ID仅仅可以做到单机程度的唯一,不利于多机扩展,可用string+hash代替

	// Type该消息元对应的对象类型
	Type reflect.Type
}

// NewMessageMeta 通过指定的id和对象类型,创建一个消息元,调用方负责保证id的唯一
func NewMessageMeta(id uint32, pType reflect.Type) *MessageMeta {
	return &MessageMeta{
		ID:   id,
		Type: pType,
	}
}

// NewType 创建一个该消息元对应的新类型对象
func (m *MessageMeta) NewType() interface{} {
	if m.Type.Kind() == reflect.Ptr {
		m.Type = m.Type.Elem()
	}

	return reflect.New(m.Type).Interface()
}
