package codec_byte

import (
	"fmt"
	"reflect"

	"github.com/MaxnSter/gnet/codec"
	"github.com/MaxnSter/gnet/iface"
)

const (
	// the name of coderByte
	CoderByteName = "byte"
)

var (
	_ iface.Coder = (*coderByte)(nil)
)

// coderByte users raw slice of bytes
type coderByte struct{}

// Encode returns raw slice of bytes
func (coder *coderByte) Encode(v interface{}) (data []byte, err error) {
	if data, ok := v.([]byte); ok {
		return data, nil
	}

	if data, ok := v.(*[]byte); ok {
		return *data, nil
	}

	return nil, fmt.Errorf("%T is not a []byte", v)
}

// Decode return raw slice of bytes
func (coder *coderByte) Decode(data []byte, v interface{}) (err error) {
	reflect.Indirect(reflect.ValueOf(v)).SetBytes(data)
	return nil
}

// return the name of coderByte
func (coder *coderByte) TypeName() string {
	return CoderByteName
}

// register coderByte
func init() {
	codec.RegisterCoder(CoderByteName, &coderByte{})
}
