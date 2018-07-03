package message_meta

// MetaIdentifier 表示一个具有唯一表示的消息对象
type MetaIdentifier interface {

	// GetId返回消息对象的唯一标识
	GetId() uint32
}
