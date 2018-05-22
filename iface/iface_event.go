package iface

type Event interface {
	Session() NetSession
	Message() interface{}
}

type MessageEvent struct {
	EventSes NetSession
	Msg      interface{}
}

func (msg *MessageEvent) Session() NetSession {
	return msg.EventSes
}

func (msg *MessageEvent) Message() interface{}{
	return msg.Msg
}

type UserEventCB interface {
	EventCB(ev Event)
}
type UserEventCBFunc func(ev Event)

func (f UserEventCBFunc) EventCB(ev Event) {
	f(ev)
}
