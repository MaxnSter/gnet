package echo

import (
	"reflect"

	"github.com/MaxnSter/gnet/example"
	"github.com/MaxnSter/gnet/message"
)

func init() {
	message.RegisterMsgMeta(example.ProtoEcho,
		message.NewMsgMeta(example.ProtoEcho, reflect.TypeOf((*EchoProto)(nil)).Elem()))
}
