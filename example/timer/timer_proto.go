package timer

import (
	"reflect"
	"time"

	"github.com/MaxnSter/gnet/example"
	"github.com/MaxnSter/gnet/message_pack/message_meta"
)

type TimerProto struct {
	Id      uint32
	TimeNow time.Time
}

func (proto *TimerProto) GetId() uint32 {
	return proto.Id
}

func init() {
	message_meta.RegisterMsgMeta(example.ProtoTimer,
		message_meta.NewMessageMeta(example.ProtoTimer, reflect.TypeOf((*TimerProto)(nil)).Elem()))
}
