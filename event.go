package gnet

// Event 是onMessage回调中传入的参数
type Event interface {
	// Session返回本条消息对应的NetSession
	Session() NetSession

	// Message返回经过unPack,decode之后的消息
	Message() interface{}
}

type eventWrapper struct {
	eventSession NetSession  //本条消息对应的NetSession
	msg          interface{} //经过UnPack,decode之后的消息
}

// Session 返回本条消息对应的NetSession
func (msg *eventWrapper) Session() NetSession {
	return msg.eventSession
}

// Message 返回经过unPack,decode之后的消息
func (msg *eventWrapper) Message() interface{} {
	return msg.msg
}
