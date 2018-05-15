package round_trip

import (
	"reflect"

	"github.com/MaxnSter/gnet/example"
	"github.com/MaxnSter/gnet/message"
)

type RoundTripProto struct {
	Id uint32
	T1 int64
	T2 int64
}

func (r *RoundTripProto) ID() uint32 {
	return r.Id
}

func init() {
	message.RegisterMsgMeta(example.ProtoRoundTrip, message.NewMsgMeta(
		example.ProtoRoundTrip, reflect.TypeOf((*RoundTripProto)(nil)).Elem(),
	))
}
