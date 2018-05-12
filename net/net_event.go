package net

import "github.com/MaxnSter/gnet/message"

type Event interface {
	Session() *TcpSession //TODO
	Message() message.Message
}

type MessageEvent struct {
	session *TcpSession
	msg     message.Message
}

func (msg *MessageEvent) Session() *TcpSession {
	return msg.session
}

func (msg *MessageEvent) Message() message.Message {
	return msg.msg
}

type UserEventCB interface {
	EventCB(ev Event)
}
type UserEventCBFunc func(ev Event)

func (f UserEventCBFunc) EventCB(ev Event) {
	f(ev)
}
