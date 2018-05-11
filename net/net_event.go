package net

import "github.com/MaxnSter/gnet/message"

type Event interface {
	Session() Session
	Message() message.Message
}

type UserEventCB interface {
	EventCB(ev Event)
}

type UserEventCBFunc func(ev Event)

func (f UserEventCBFunc) EventCB(ev Event) {
	f(ev)
}
