package gnet

import (
	"io"

	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/message_pack/message_meta"
)

// Operator 可以理解网络组件和module的中间件.
// 网络组件提供读写对象,module提供对消息流的"解释"(封解包,序列化,反序列化).
// 同时,module无需知道读写对象是谁,也根本不知道网络组件的存在.
// Operator负责从reader提供方读数据,使用module定义的"解释"方式,最后传入业务逻辑方.写操作同理
type Operator interface {
	// StartModule启动指定module中的所有组件
	StartModule(module Module)
	// StopModule停止指定module中的组件,调用法阻塞直到module完全关闭
	StopModule(module Module)
	// StopModuleAsync停止指定module中的组件并立即返回
	StopModuleAsync(module Module) []<-chan struct{}

	// PostEvent将Event派发至指定module的goroutine pool中
	PostEvent(ev Event, module Module)

	// Read从网络组件提供的reader中读数据,并从使用module提供的packer和coder的到一个完整的消息对象
	Read(reader io.Reader, module Module) (interface{}, error)
	// Write使用module提供的packer和coder得到最终写入数据,并从网络组件提供的writer中写数据
	Write(writer io.Writer, msg interface{}, module Module) error

	//TODO 从Operator的作用来看,回调作为module组件似乎更合理
	// SetOnMessage设置消息接收的回调
	SetOnMessage(onMessage OnMessage)
	// SetOnConnected设置连接建立的回调
	SetOnConnected(onConnected OnConnected)
	// SetOnClose设置连接关闭的回调
	SetOnClose(onClose OnClose)

	// GetOnMessage返回已注册的消息接收回调
	GetOnMessage() OnMessage
	// GetOnConnected返回已注册的连接建立回调
	GetOnConnected() OnConnected
	// GetOnClose返回已注册的连接关闭回调
	GetOnClose() OnClose
}

// OnMessage 为接收消息的回调
type OnMessage func(ev Event)

// OnConnected 为连接建立的回调
type OnConnected func(session NetSession)

// OnClose 为连接关闭的回调
type OnClose func(session NetSession)

// NewOperator 通过指定的消息接收回调,创建一个Operator
func NewOperator(cb OnMessage) Operator {
	o := &operatorWrapper{}
	o.SetOnMessage(cb)
	return o
}

type operatorWrapper struct {
	OnMessageFunc   OnMessage
	OnConnectedFunc OnConnected
	OnCloseFunc     OnClose
}

// StartModule启动指定module中的所有组件
func (o *operatorWrapper) StartModule(m Module) {
	if m.Pool() != nil {
		m.Pool().Start()
	}

	if m.Timer() != nil {
		m.Timer().Start()
	}
}

// StopModule停止指定module中的组件,调用法阻塞直到module完全关闭
func (o *operatorWrapper) StopModule(m Module) {
	if m.Timer() != nil {
		m.Timer().Stop()
	}

	if m.Pool() != nil {
		m.Pool().Stop()
	}
}

// StopModuleAsync停止指定module中的组件并立即返回
func (o *operatorWrapper) StopModuleAsync(m Module) (stops []<-chan struct{}) {
	if m.Timer() != nil {
		stops = append(stops, m.Timer().StopAsync())
	}

	if m.Pool() != nil {
		stops = append(stops, m.Pool().StopAsync())
	}
	return
}

// PostEvent将Event派发至指定module的goroutine pool中
func (o *operatorWrapper) PostEvent(ev Event, module Module) {
	if module.Pool() == nil {
		o.OnMessageFunc(ev)
		return
	}
	module.Pool().Put(ev.Session(), func(_ iface.Context) {
		o.OnMessageFunc(ev)
	})
}

// Read从网络组件提供的reader中读数据,并从使用module提供的packer和coder的到一个完整的消息对象
func (o *operatorWrapper) Read(reader io.Reader, module Module) (interface{}, error) {
	var meta *message_meta.MessageMeta
	plugins := module.RdPlugins()
	c := module.Coder()

	if len(plugins) > 0 {
		for _, plugin := range plugins {
			reader, c, meta = plugin.BeforeRead(reader, c, meta)
		}
	}

	return module.Packer().Unpack(reader, c, meta)
}

// Write使用module提供的packer和coder得到最终写入数据,并从网络组件提供的writer中写数据
func (o *operatorWrapper) Write(writer io.Writer, msg interface{}, module Module) error {
	plugins := module.WrPlugins()
	c := module.Coder()

	if len(plugins) > 0 {
		for _, plugin := range plugins {
			writer, c, msg = plugin.BeforeWrite(writer, c, msg)
		}
	}

	return module.Packer().Pack(writer, c, msg)
}

// SetOnMessage设置接收的回调
func (o *operatorWrapper) SetOnMessage(onMessage OnMessage) {
	o.OnMessageFunc = onMessage
}

// GetOnMessage返回已注册的消息接收回调
func (o *operatorWrapper) GetOnMessage() OnMessage {
	return o.OnMessageFunc
}

// SetOnConnected设置连接建立的回调
func (o *operatorWrapper) SetOnConnected(onConnected OnConnected) {
	o.OnConnectedFunc = onConnected
}

// GetOnConnected返回已注册的连接建立回调
func (o *operatorWrapper) GetOnConnected() OnConnected {
	return o.OnConnectedFunc
}

// SetOnClose设置连接关闭的回调
func (o *operatorWrapper) SetOnClose(onClose OnClose) {
	o.OnCloseFunc = onClose
}

// GetOnClose返回已注册的连接关闭回调
func (o *operatorWrapper) GetOnClose() OnClose {
	return o.OnCloseFunc
}
