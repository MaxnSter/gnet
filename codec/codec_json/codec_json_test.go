package codec_json

import (
	"testing"

	"github.com/MaxnSter/gnet/codec"
	"github.com/stretchr/testify/assert"
)

func TestEncodeAndDecode(t *testing.T) {

	type Info struct {
		Id  uint32
		Msg string
	}

	coder := codec.MustGetCoder("json")

	data, err := coder.Encode(&Info{Id: 1, Msg: "json"})
	assert.Nil(t, err, err)

	newInfo := new(Info)
	err = coder.Decode(data, newInfo)
	assert.Nil(t, err, err)
	assert.NotNil(t, newInfo)

	assert.Equal(t, "json", newInfo.Msg)
	assert.Equal(t, uint32(1), newInfo.Id)

}
