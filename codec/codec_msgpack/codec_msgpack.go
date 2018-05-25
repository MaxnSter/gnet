package codec_msgpack

import (
	"github.com/MaxnSter/gnet/codec"
	"github.com/MaxnSter/gnet/iface"
	"github.com/vmihailenco/msgpack"
)

const (
	// the name of coderMsgpack
	CoderMsgPackTypeName = "msgpack"
)

var (
	_ iface.Coder = (*coderMsgpack)(nil)
)

// coderMsgpack uses messagepack marshaler and unmarshaller
type coderMsgpack struct{}

// return the name of coderMsgpack
func (c coderMsgpack) TypeName() string {
	return CoderMsgPackTypeName
}

// Encode encodes an object into slice of bytes
func (c coderMsgpack) Encode(msg interface{}) (data []byte, err error) {
	return msgpack.Marshal(msg)
}

// Decode decodes an object from slice of bytes
func (c coderMsgpack) Decode(data []byte, pMsg interface{}) error {
	return msgpack.Unmarshal(data, pMsg)
}

// register coderMsgpack
func init() {
	codec.RegisterCoder(CoderMsgPackTypeName, &coderMsgpack{})
}
