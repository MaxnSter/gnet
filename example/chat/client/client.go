package client

import (
	"github.com/MaxnSter/gnet"
	_ "github.com/MaxnSter/gnet/codec/codec_byte"
	"github.com/MaxnSter/gnet/iface"
	_ "github.com/MaxnSter/gnet/message_pack/pack/pack_length_value"
	"github.com/MaxnSter/gnet/net"
	_ "github.com/MaxnSter/gnet/worker_pool/worker_session_race_other"
)

type chatClient struct {
	*net.TcpClient
	*net.TcpSession

	addr string

	connectedHook func(session *net.TcpSession)
	onMessageHook func(ev iface.Event)
}

func NewChatClient(addr string, connHook func(*net.TcpSession), msgHook func(iface.Event)) *chatClient {
	c := &chatClient{addr: addr, connectedHook: connHook, onMessageHook: msgHook}

	// build a tcp client
	option := &gnet.GnetOption{Coder: "byte", Packer: "lv", WorkerPool: "poolRaceOther"}
	callback := gnet.NewCallBackOption(gnet.WithOnConnectCB(c.onConnect))
	c.TcpClient = gnet.NewClient(c.addr, callback, option, c.onMessage)

	return c
}

func NewChatClientWithPool(addr string, connHook func(*net.TcpSession), msgHook func(iface.Event),
	pool iface.WorkerPool) *chatClient {

	c := &chatClient{addr: addr, connectedHook: connHook, onMessageHook: msgHook}
	// build a tcp client
	option := &gnet.GnetOption{Coder: "byte", Packer: "lv", WorkerPool: "poolRaceOther"}
	callback := gnet.NewCallBackOption(gnet.WithOnConnectCB(c.onConnect))
	c.TcpClient = gnet.NewClientSharePool(c.addr, callback, option, pool, c.onMessage)

	return c
}

func (c *chatClient) Run() {
	c.StartAndRun()
}

func (c *chatClient) onConnect(session *net.TcpSession) {
	c.TcpSession = session
	if c.connectedHook != nil {
		c.connectedHook(c.TcpSession)
	}
}

func (c *chatClient) onMessage(ev iface.Event) {
	switch ev.Message().(type) {
	case []byte:
		if c.onMessageHook != nil {
			c.onMessageHook(ev)
		}
	default:
		//error msg
	}
}
