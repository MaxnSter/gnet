package gnet

import (
	"io"

	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/message_pack/message_meta"
)

type Operator interface {
	StartModule(module Module)
	StopModule(module Module)
	StopModuleAsync(module Module) []<-chan struct{}

	PostEvent(ev Event, module Module)
	Read(reader io.Reader, module Module) (interface{}, error)
	Write(writer io.Writer, msg interface{}, module Module) error

	SetOnMessage(onMessage OnMessage)
	GetOnMessage() OnMessage

	SetOnConnected(onConnected OnConnected)
	GetOnConnected() OnConnected

	SetOnClose(onClose OnClose)
	GetOnClose() OnClose
}

type OnMessage func(ev Event)
type OnConnected func(session NetSession)
type OnClose func(session NetSession)

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

func (o *operatorWrapper) StartModule(m Module) {
	if m.Pool() != nil {
		m.Pool().Start()
	}

	if m.Timer() != nil {
		m.Timer().Start()
	}
}

func (o *operatorWrapper) StopModule(m Module) {
	if m.Timer() != nil {
		m.Timer().Stop()
	}

	if m.Pool() != nil {
		m.Pool().Stop()
	}
}

func (o *operatorWrapper) StopModuleAsync(m Module) (stops []<-chan struct{}) {
	if m.Timer() != nil {
		stops = append(stops, m.Timer().StopAsync())
	}

	if m.Pool() != nil {
		stops = append(stops, m.Pool().StopAsync())
	}
	return
}

func (o *operatorWrapper) PostEvent(ev Event, module Module) {
	if module.Pool() == nil {
		o.OnMessageFunc(ev)
		return
	}
	module.Pool().Put(ev.Session(), func(_ iface.Context) {
		o.OnMessageFunc(ev)
	})
}

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

func (o *operatorWrapper) SetOnMessage(onMessage OnMessage) {
	o.OnMessageFunc = onMessage
}

func (o *operatorWrapper) GetOnMessage() OnMessage {
	return o.OnMessageFunc
}

func (o *operatorWrapper) SetOnConnected(onConnected OnConnected) {
	o.OnConnectedFunc = onConnected
}

func (o *operatorWrapper) GetOnConnected() OnConnected {
	return o.OnConnectedFunc
}

func (o *operatorWrapper) SetOnClose(onClose OnClose) {
	o.OnCloseFunc = onClose
}

func (o *operatorWrapper) GetOnClose() OnClose {
	return o.OnCloseFunc
}
