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

func (msg *MessageEvent) Message() interface{} {
	return msg.Msg
}

type OnMessage interface {
	OnMessageCB(ev Event)
}

type OnMessageFunc func(ev Event)

func (f OnMessageFunc) OnMessageCB(ev Event) {
	f(ev)
}
