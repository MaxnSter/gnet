package codec_json

import (
	"encoding/json"

	"github.com/MaxnSter/gnet/codec"
)

var (
	_ codec.Coder = (*coderJson)(nil)
)

const (
	// the name of coderJson
	CoderJsonTypeName = "json"
)

// coderJson use json marshaler and unmarshaler
type coderJson struct{}

// return the name of coderJson
func (c *coderJson) TypeName() string {
	return CoderJsonTypeName
}

// Encode encodes an object into slice of bytes
func (c *coderJson) Encode(v interface{}) (data []byte, err error) {
	return json.Marshal(v)
}

// Decode decodes an object from from slice of bytes
func (c *coderJson) Decode(data []byte, v interface{}) (err error) {
	return json.Unmarshal(data, v)
}

// register coderJson
func init() {
	c := &coderJson{}
	codec.RegisterCoder(c.TypeName(), c)
}
