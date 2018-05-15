package echo

import (
	"reflect"

	"github.com/MaxnSter/gnet/example"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/message"
)

var (
	_ iface.Message = (*DemoMessage)(nil)
)

const (
	protocolId = example.ProtoDemoMessage
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
