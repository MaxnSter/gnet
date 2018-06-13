package gnet

type Event interface {
	Session() NetSession
	Message() interface{}
}

type EventWrapper struct {
	EventSession NetSession
	Msg          interface{}
}

func (msg *EventWrapper) Session() NetSession {
	return msg.EventSession
}

func (msg *EventWrapper) Message() interface{} {
	return msg.Msg
}

