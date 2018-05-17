package codec_msgpack

import (
	"github.com/MaxnSter/gnet/codec"
	"github.com/MaxnSter/gnet/iface"
	"github.com/vmihailenco/msgpack"
)

const (
	CoderMsgPackTypeName = "msgpack"
)

var (
	_ iface.Coder = (*CoderMsgpack)(nil)
)

type CoderMsgpack struct {
}

func (c CoderMsgpack) TypeName() string {
	return CoderMsgPackTypeName
}

func (c CoderMsgpack) Encode(msg interface{}) (data []byte, err error) {
	return msgpack.Marshal(msg)
}

func (c CoderMsgpack) Decode(data []byte, pMsg interface{}) error {
	return msgpack.Unmarshal(data, pMsg)
}

func init() {
	codec.RegisterCoder(CoderMsgPackTypeName, &CoderMsgpack{})
}
