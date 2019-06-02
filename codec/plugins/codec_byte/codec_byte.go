package codec_byte

import (
	"fmt"
	"reflect"

	"github.com/MaxnSter/gnet/codec"
)

const (
	// the name of coderByte
	coderByteName = "byte"
)

var (
	_ codec.Coder = (*coderByte)(nil)
)

// coderByte users raw slice of bytes
type coderByte struct{}

// Encode returns raw slice of bytes
func (coder *coderByte) Encode(v interface{}) (data []byte, err error) {
	if v, ok := v.([]byte); ok {
		data = make([]byte, len(v))
		copy(data, v)
		return
	}

	if v, ok := v.(*[]byte); ok {
		data = make([]byte, len(*v))
		copy(data, *v)
		return
	}

	if v, ok := v.(string); ok {
		data = []byte(v)
		return
	}

	if v, ok := v.(*string); ok {
		data = []byte(*v)
		return
	}

	return nil, fmt.Errorf("%T is not a []byte or string", v)
}

// Decode return raw slice of bytes
func (coder *coderByte) Decode(data []byte, v interface{}) (err error) {

	if _, ok := v.(*[]byte); !ok {
		return fmt.Errorf("%T is not a *[]byte", v)
	}

	reflect.Indirect(reflect.ValueOf(v)).SetBytes(data)
	return nil
}

// return the name of coderByte
func (coder *coderByte) String() string {
	return coderByteName
}

// register coderByte
func init() {
	codec.RegisterCoder(coderByteName, &coderByte{})
}

func New() codec.Coder {
	return &coderByte{}
}
