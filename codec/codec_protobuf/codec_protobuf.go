package codec_protobuf

import (
	"errors"

	"github.com/MaxnSter/gnet/codec"
	"github.com/golang/protobuf/proto"
)

const (
	// the name of coderProtobuf
	CoderProtoTypeName = "protobuf"
)

var (
	_ codec.Coder = (*coderProtobuf)(nil)
)

// coderProtobuf uses protobuf marshaller and unmarshaller
type coderProtobuf struct{}

// return the name of coderProtobuf
func (p *coderProtobuf) TypeName() string {
	return CoderProtoTypeName
}

// Encode encodes an object into slice of bytes
func (p *coderProtobuf) Encode(msg interface{}) (data []byte, err error) {
	if protoMsg, ok := msg.(proto.Message); ok {
		return proto.Marshal(protoMsg)
	} else {
		//TODO errors
		return nil, errors.New("type assert error")
	}
}

// Decode decodes an object from slice of bytes
func (p *coderProtobuf) Decode(data []byte, pMsg interface{}) error {
	if protoMsg, ok := pMsg.(proto.Message); ok {
		return proto.Unmarshal(data, protoMsg)
	} else {
		//TODO errors
		return errors.New("type assert error")
	}
}

// register coderProtobuf
func init() {
	codec.RegisterCoder(CoderProtoTypeName, &coderProtobuf{})
}
