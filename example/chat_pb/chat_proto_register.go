package chat_pb

import (
	"reflect"

	"github.com/MaxnSter/gnet/message_pack/message_meta"
)

const (
	// 消息id保证唯一
	ChatMsgId = 0
)

func init() {
	// 注册该id对应的类型信息,我们使用tlv Packer时可以根据解析的id反射生成对应类型的struct
	metaType := reflect.TypeOf((*ChatMessage)(nil)).Elem()
	message_meta.RegisterMsgMeta(&message_meta.MessageMeta{ChatMsgId, metaType})
}
