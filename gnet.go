package gnet

import (
	"github.com/MaxnSter/gnet/codec"
	_ "github.com/MaxnSter/gnet/codec/codec_msgpack"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/net"
	"github.com/MaxnSter/gnet/message_pack"
	_ "github.com/MaxnSter/gnet/message_pack/pack/pack_type_length_value"
	"github.com/MaxnSter/gnet/timer"
	"github.com/MaxnSter/gnet/worker"
	_ "github.com/MaxnSter/gnet/worker/worker_session_race_self"
)

//TODO unp, http...
func NewServer(addr string,
	cbOption *CallBackOption,
	gnetOption *GnetOption,
	onMessage iface.OnMessageFunc) *net.TcpServer {

	netOp := newNetOption(cbOption, gnetOption, onMessage)
	return net.NewTcpServer(addr, "", netOp)
}

func NewServerSharePool(addr string,
	cbOption *CallBackOption,
	gnetOption *GnetOption,
	sharePool iface.WorkerPool,
	onMessage iface.OnMessageFunc) *net.TcpServer {

	netOp := &net.NetOptions{
		Coder:  codec.MustGetCoder(gnetOption.Coder),
		Pool:   sharePool,
		Packer: message_pack.MustGetPacker(gnetOption.Packer),
		CB:     onMessage,

		OnConnected:    cbOption.OnConnect,
		OnSessionClose: cbOption.OnSessionClose,
		OnServerClosed: cbOption.OnServerClosed,
	}
	netOp.Timer = timer.NewTimerManager(netOp.Pool)

	return net.NewTcpServer(addr, "", netOp)
}

func NewClient(addr string,
	cbOption *CallBackOption,
	gnetOption *GnetOption,
	onMessage iface.OnMessageFunc) *net.TcpClient {

	netOp := newNetOption(cbOption, gnetOption, onMessage)
	return net.NewTcpClient(addr, netOp)
}

func NewClientSharePool(addr string,
	cbOption *CallBackOption,
	gnetOption *GnetOption,
	sharePool iface.WorkerPool,
	onMessage iface.OnMessageFunc) *net.TcpClient {

	netOp := &net.NetOptions{
		Coder:  codec.MustGetCoder(gnetOption.Coder),
		Pool:   sharePool,
		Packer: message_pack.MustGetPacker(gnetOption.Packer),
		CB:     onMessage,

		OnConnected:    cbOption.OnConnect,
		OnSessionClose: cbOption.OnSessionClose,
		OnServerClosed: cbOption.OnServerClosed,
	}
	netOp.Timer = timer.NewTimerManager(netOp.Pool)

	return net.NewTcpClient(addr, netOp)
}

func newNetOption(cbOption *CallBackOption, gnetOption *GnetOption, onMessage iface.OnMessage) *net.NetOptions {

	netOp := &net.NetOptions{
		Coder:  codec.MustGetCoder(gnetOption.Coder),
		Pool:   worker.MustGetWorkerPool(gnetOption.WorkerPool),
		Packer: message_pack.MustGetPacker(gnetOption.Packer),
		CB:     onMessage,

		OnConnected:    cbOption.OnConnect,
		OnSessionClose: cbOption.OnSessionClose,
		OnServerClosed: cbOption.OnServerClosed,
	}
	netOp.Timer = timer.NewTimerManager(netOp.Pool)

	return netOp
}

var defaultCbOption = &CallBackOption{}
var defaultGnetOption = &GnetOption{Packer: "tlv", Coder: "msgpack", WorkerPool: "poolRaceSelf"}

type CallBackOption struct {
	OnConnect      net.OnConnectedFunc
	OnSessionClose net.OnSessionCloseFunc
	OnServerClosed net.OnServerClosedFunc
}

func WithOnConnectCB(onConnect net.OnConnectedFunc) func(*CallBackOption) {
	return func(o *CallBackOption) {
		o.OnConnect = onConnect
	}
}

func WithOnSessionClose(onClose net.OnSessionCloseFunc) func(*CallBackOption) {
	return func(o *CallBackOption) {
		o.OnSessionClose = onClose
	}
}

func WithOnServerClosed(onClosed net.OnServerClosedFunc) func(*CallBackOption) {
	return func(o *CallBackOption) {
		o.OnServerClosed = onClosed
	}
}

func NewCallBackOption(options ...func(*CallBackOption)) *CallBackOption {
	o := *defaultCbOption
	for _, option := range options {
		option(&o)
	}

	return &o
}

type GnetOption struct {
	Packer     string
	Coder      string
	WorkerPool string
}

func NewGnetOption(options ...func(*GnetOption)) *GnetOption {
	o := *defaultGnetOption
	for _, option := range options {
		option(&o)
	}

	return &o
}

func WithPacker(packer string) func(*GnetOption) {
	return func(o *GnetOption) {
		o.Packer = packer
	}
}

func WithCoder(coder string) func(*GnetOption) {
	return func(o *GnetOption) {
		o.Coder = coder
	}
}

func WithWorkerPool(workerPool string) func(*GnetOption) {
	return func(o *GnetOption) {
		o.WorkerPool = workerPool
	}
}
