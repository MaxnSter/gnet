package message_demo

import (
	"reflect"

	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/message"
	"github.com/MaxnSter/gnet/message/protocol"
)

var (
	_ iface.Message = (*DemoMessage)(nil)
)

const (
	protocolId = protocol.ProtoDemoMessage
)

type DemoMessage struct {
	Id  uint32
	Val string
}

func NewDemoMessage(v string) *DemoMessage {
	return &DemoMessage{
		Id:  protocolId,
		Val: v,
	}
}

func (msg *DemoMessage) ID() uint32 {
	return msg.Id
}

func init() {
	meta := message.NewMsgMeta(protocolId, reflect.TypeOf((*DemoMessage)(nil)).Elem())
	message.RegisterMsgMeta(protocolId, meta)
}
