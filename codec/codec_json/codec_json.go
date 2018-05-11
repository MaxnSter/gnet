package codec_json

import (
	"encoding/json"

	"gnet/codec"
)

var (
	_ codec.Coder = (*coderJson)(nil)
)

const (
	// codec_json's typeName
	CoderJsonTypeName = "json"
)

type coderJson struct {
}

func (c *coderJson) TypeName() string {
	return CoderJsonTypeName
}

func (c *coderJson) Encode(v interface{}) (data []byte, err error) {
	return json.Marshal(v)
}

func (c *coderJson) Decode(data []byte, v interface{}) (err error) {
	return json.Unmarshal(data, v)
}

func init() {
	c := &coderJson{}
	codec.RegisterCoder(c.TypeName(), c)
}
