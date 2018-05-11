package gnet

import (
	"github.com/MaxnSter/gnet/codec"
	"github.com/MaxnSter/gnet/codec/codec_json"
	"github.com/MaxnSter/gnet/net"
	"github.com/MaxnSter/gnet/pack"
	"github.com/MaxnSter/gnet/pack/pack_type_length_value"
)

const (
	defaultCoder  = codec_json.CoderJsonTypeName
	defaultPacker = pack_type_length_value.TlvPackerName
)

func WithCoder(coderName string) net.NetOpFunc {

	return func(options *net.NetOptions) {
		options.Coder = codec.MustGetCoder(coderName)
	}
}

func WithPacker(packerName string) net.NetOpFunc {
	return func(options *net.NetOptions) {
		options.Packer = pack.MustGetPacker(packerName)
	}
}

//TODO change return
func NewServer(netWork string, addr string, options ...net.NetOpFunc) *net.TcpServer {
	return &net.TcpServer{}
}
