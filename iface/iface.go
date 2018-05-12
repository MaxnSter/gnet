package iface

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

func WithConnectedCB(cb net.OnConnectedFunc) net.NetOpFunc {
	return func(options *net.NetOptions) {
		options.OnConnect = cb
	}
}

func WithCloseCB(cb net.OnSessionCloseFunc) net.NetOpFunc {
	return func(options *net.NetOptions) {
		options.OnClose = cb
	}
}

func WithServerCloseCB(cb net.OnServerClosedFunc) net.NetOpFunc {
	return func(options *net.NetOptions) {
		options.OnServerClose = cb
	}
}

func WithAcceptCB(cb net.OnAcceptedFunc) net.NetOpFunc {
	return func(options *net.NetOptions) {
		options.OnAccepted = cb
	}
}

//TODO change return
func NewServer(addr string, name string, cb net.UserEventCBFunc, options ...net.NetOpFunc) *net.TcpServer {
	op := net.NetOptions{
		Packer: pack.MustGetPacker(defaultPacker),
		Coder:  codec.MustGetCoder(defaultCoder),
	}
	for _, f := range options {
		f(&op)
	}
	op.CB = cb

	return net.NewTcpServer(addr, name, op)
}

func NewClient(addr string, cb net.UserEventCBFunc, options ...net.NetOpFunc) *net.TcpClient {
	op := net.NetOptions{
		Packer: pack.MustGetPacker(defaultPacker),
		Coder:  codec.MustGetCoder(defaultCoder),
	}
	for _, f := range options {
		f(&op)
	}
	op.CB = cb

	return net.NewTcpClient(addr, op)
}
