package net

import (
	"io"

	"github.com/MaxnSter/gnet/iface"
)

type NetOptions struct {
	Coder  iface.Coder
	Packer iface.Packer
	Worker iface.WorkerPool
	CB     iface.UserEventCB

	OnConnected    OnConnectedFunc
	OnClose        OnSessionCloseFunc
	OnServerClosed OnServerClosedFunc

	//TODO more, such as socket options...
}

type OnConnectedFunc func(session *TcpSession)
type OnSessionCloseFunc func(session *TcpSession)
type OnServerClosedFunc func()

type NetOpFunc func(options *NetOptions)

func (op *NetOptions) ReadMessage(reader io.Reader) (iface.Message, error) {
	return op.Packer.Unpack(reader, op.Coder)
}

func (op *NetOptions) WriteMessage(writer io.Writer, msg iface.Message) error {
	return op.Packer.Pack(writer, op.Coder, msg)
}

func (op *NetOptions) PostEvent(ev iface.Event) {
	op.Worker.Put(ev.Session(), func() {
		op.CB.EventCB(ev)
	})
}
