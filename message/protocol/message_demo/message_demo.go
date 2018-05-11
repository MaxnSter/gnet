package message_demo

import (
	"reflect"

	"gnet/message"
	"gnet/message/protocol"
)

var (
	_ message.Message = (*DemoMessage)(nil)
)

const (
	protocolId = protocol.ProtoDemoMessage
)

type DemoMessage struct {
	id  uint32
	Val string
}

func NewDemoMessage(v string) *DemoMessage {
	return &DemoMessage{
		id:  protocolId,
		Val: v,
	}
}

func (msg *DemoMessage) ID() uint32 {
	return msg.id
}

func init() {
	message.RegisterMsgMeta(protocolId, &message.MessageMeta{
		ID:   protocolId,
		Type: reflect.TypeOf((*DemoMessage)(nil)).Elem(),
	})
}
