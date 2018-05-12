package net

import (
	"io"

	"github.com/MaxnSter/gnet/codec"
	"github.com/MaxnSter/gnet/message"
	"github.com/MaxnSter/gnet/pack"
)

type NetOptions struct {
	Coder  codec.Coder
	Packer pack.Packer
	CB     UserEventCB

	OnConnect     OnConnectedFunc
	OnAccepted    OnAcceptedFunc
	OnClose       OnSessionCloseFunc
	OnServerClose OnServerClosedFunc

	//TODO more, such as socket options...
}

type OnConnectedFunc func(session *TcpSession)
type OnAcceptedFunc func(session *TcpSession)
type OnSessionCloseFunc func(session *TcpSession)
type OnServerClosedFunc func()

type NetOpFunc func(options *NetOptions)

func (op *NetOptions) ReadMessage(reader io.Reader) (message.Message, error) {
	return op.Packer.Unpack(reader, op.Coder)
}

func (op *NetOptions) WriteMessage(writer io.Writer, msg message.Message) error {
	return op.Packer.Pack(writer, op.Coder, msg)
}

func (op *NetOptions) PostEvent(ev Event) {
	op.CB.EventCB(ev)
}
