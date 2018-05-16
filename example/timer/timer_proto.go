package timer

import (
	"reflect"
	"time"

	"github.com/MaxnSter/gnet/example"
	"github.com/MaxnSter/gnet/message"
)

type TimerProto struct {
	Id      uint32
	TimeNow time.Time
}

func (proto *TimerProto) GetId() uint32 {
	return proto.Id
}

func init() {
	message.RegisterMsgMeta(example.ProtoTimer,
		message.NewMsgMeta(example.ProtoTimer, reflect.TypeOf((*TimerProto)(nil)).Elem()))
}
