package test

import (
	"testing"

	"github.com/MaxnSter/gnet/codec"
	"github.com/stretchr/testify/assert"

	_ "github.com/MaxnSter/gnet/codec/codec_protobuf"
)

func TestEncodeAndDecode(t *testing.T) {
	coder := codec.MustGetCoder("protoBuf")

	data, err := coder.Encode(&Info{Id: 1, Msg: "proto"})
	assert.Nil(t, err, err)

	newInfo := new(Info)
	err = coder.Decode(data, newInfo)
	assert.Nil(t, err, err)
	assert.NotNil(t, newInfo)

	assert.Equal(t, "proto", newInfo.Msg)
	assert.Equal(t, uint32(1), newInfo.Id)

}
