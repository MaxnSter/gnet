package message_pack

import (
	"io"

	"github.com/MaxnSter/gnet/codec"
	"github.com/MaxnSter/gnet/message_pack/message_meta"
)

// Packer 定义了一个封包解包器
type Packer interface {

	// Unpack从reader中读消息并解包,然后通过指定的coder和meta序列化出msg对象
	// 注意:gnet中传入的meta默认为nil,因为我们根本无法知道接收的消息类型
	// 解决方案1:若事先约定好一种消息类型且不会变,可添加readPlugin插件,手动传入约定的meta
	// 解决方案2:使用gnet内置的tlv封包解包器,可以避免上述问题
	Unpack(reader io.Reader, c codec.Coder, meta *message_meta.MessageMeta) (msg interface{}, err error)

	// Pack使用指定的coder序列化msg然后封包,最后写入writer
	Pack(writer io.Writer, c codec.Coder, msg interface{}) error

	// TypeName返回该packer的名称
	TypeName() string
}
