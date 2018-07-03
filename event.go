package gnet

// Event 是onMessage回调中传入的参数
type Event interface {
	// Session返回本条消息对应的NetSession
	Session() NetSession

	// Message返回经过unPack,decode之后的消息
	Message() interface{}
}

// EventWrapper 是Event的一个实现
type EventWrapper struct {
	EventSession NetSession  //本条消息对应的NetSession
	Msg          interface{} //经过UnPack,decode之后的消息
}

// Session 返回本条消息对应的NetSession
func (msg *EventWrapper) Session() NetSession {
	return msg.EventSession
}

// Message 返回经过unPack,decode之后的消息
func (msg *EventWrapper) Message() interface{} {
	return msg.Msg
}
