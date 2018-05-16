package gnet

import (
	"github.com/MaxnSter/gnet/codec"
	_ "github.com/MaxnSter/gnet/codec/codec_json"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/net"
	"github.com/MaxnSter/gnet/pack"
	_ "github.com/MaxnSter/gnet/pack/pack_type_length_value"
	"github.com/MaxnSter/gnet/timer"
	"github.com/MaxnSter/gnet/worker"
	_ "github.com/MaxnSter/gnet/worker/worker_session_race_other"
	_ "github.com/MaxnSter/gnet/worker/worker_session_race_self"
)

const (
	defaultCoder            = "json"
	defaultPacker           = "tlv"
	defaultServerWorkerPool = "poolRaceSelf"
	defaultClientWorkerPool = "poolRaceOther"
)

func WithWorkerPool(poolName string) net.NetOpFunc {
	return func(options *net.NetOptions) {
		options.Worker = worker.MustGetWorkerPool(poolName)
	}
}

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
		options.OnConnected = cb
	}
}

func WithCloseCB(cb net.OnSessionCloseFunc) net.NetOpFunc {
	return func(options *net.NetOptions) {
		options.OnClose = cb
	}
}

func WithServerCloseCB(cb net.OnServerClosedFunc) net.NetOpFunc {
	return func(options *net.NetOptions) {
		options.OnServerClosed = cb
	}
}

func getDefaultOptions() net.NetOptions {
	return net.NetOptions{
		Packer: pack.MustGetPacker(defaultPacker),
		Coder:  codec.MustGetCoder(defaultCoder),
		Worker: worker.MustGetWorkerPool(defaultServerWorkerPool),
	}
}

//TODO unp, http...
func NewServer(addr string, name string, cb iface.UserEventCBFunc, options ...net.NetOpFunc) *net.TcpServer {
	op := getDefaultOptions()
	for _, f := range options {
		f(&op)
	}
	op.CB = cb
	op.Timer = timer.NewTimerManager(op.Worker)

	return net.NewTcpServer(addr, name, &op)
}

func NewClient(addr string, cb iface.UserEventCBFunc, options ...net.NetOpFunc) *net.TcpClient {
	op := getDefaultOptions()
	WithWorkerPool(defaultClientWorkerPool)(&op)
	for _, f := range options {
		f(&op)
	}
	op.CB = cb
	op.Timer = timer.NewTimerManager(op.Worker)

	return net.NewTcpClient(addr, &op)
}
