package codec_msgpack

import (
	"github.com/MaxnSter/gnet/codec"
	"github.com/vmihailenco/msgpack"
)

const (
	// the name of coderMsgpack
	coderMsgPackTypeName = "msgpack"
)

var (
	_ codec.Coder = (*coderMsgpack)(nil)
)

// coderMsgpack uses messagepack marshaler and unmarshaller
type coderMsgpack struct{}

// return the name of coderMsgpack
func (c coderMsgpack) TypeName() string {
	return coderMsgPackTypeName
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
	codec.RegisterCoder(coderMsgPackTypeName, &coderMsgpack{})
}
