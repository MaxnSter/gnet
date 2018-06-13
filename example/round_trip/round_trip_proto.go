package round_trip

import (
	"reflect"

	"github.com/MaxnSter/gnet/example"
	"github.com/MaxnSter/gnet/message_pack/message_meta"
)

type RoundTripProto struct {
	Id uint32
	T1 int64
	T2 int64
}

func (r *RoundTripProto) GetId() uint32 {
	return r.Id
}

func init() {
	message_meta.RegisterMsgMeta(example.ProtoRoundTrip, message_meta.NewMessageMeta(
		example.ProtoRoundTrip, reflect.TypeOf((*RoundTripProto)(nil)).Elem(),
	))
}
