package codec_msgpack

import (
	"testing"

	"github.com/MaxnSter/gnet/codec"
	"github.com/stretchr/testify/assert"
)

func TestCoderMsgpack_DecodeAndEchode(t *testing.T) {
	type Info struct {
		Id  uint32
		Msg string
	}

	coder := codec.MustGetCoder("msgpack")
	data, err := coder.Encode(&Info{Id: 1, Msg: "msgpack"})
	assert.Nil(t, err, err)

	newInfo := new(Info)
	err = coder.Decode(data, newInfo)
	assert.Nil(t, err, err)
	assert.NotNil(t, newInfo)

	assert.Equal(t, uint32(1), newInfo.Id)
	assert.Equal(t, "msgpack", newInfo.Msg)
}
