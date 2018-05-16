package codec_protobuf

import (
	"github.com/MaxnSter/gnet/codec"
	"github.com/MaxnSter/gnet/iface"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

const (
	CoderProtoTypeName = "protoBuf"
)

var (
	_ iface.Coder = (*coderProtobuf)(nil)
)

type coderProtobuf struct {
}

func (p *coderProtobuf) TypeName() string {
	return CoderProtoTypeName
}

func (p *coderProtobuf) Encode(msg interface{}) (data []byte, err error) {
	if protoMsg, ok := msg.(proto.Message); ok {
		return proto.Marshal(protoMsg)
	} else {
		//TODO errors
		return nil, errors.New("type assert error")
	}
}

func (p *coderProtobuf) Decode(data []byte, pMsg interface{}) error {
	if protoMsg, ok := pMsg.(proto.Message); ok {
		return proto.Unmarshal(data, protoMsg)
	} else {
		return errors.New("type assert error")
	}
}

func init() {
	codec.RegisterCoder(CoderProtoTypeName, &coderProtobuf{})
}
