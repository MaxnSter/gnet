package plugin

import (
	"io"

	"github.com/MaxnSter/gnet/codec"
	"github.com/MaxnSter/gnet/message_pack/message_meta"
)

// PluginBeforeRead 为socket读操作的hook
// 流程为:netSession start read -> PluginBeforeRead... -> read from socket
type PluginBeforeRead interface {
	// BeforeRead在真正执行socket读操作前调用
	// 通过添加插件,调用方可以改变即将使用的coder,meta,甚至改变reader(不推荐)
	BeforeRead(io.Reader, codec.Coder, *message_meta.MessageMeta) (io.Reader, codec.Coder, *message_meta.MessageMeta)
}

// PluginBeforeWrite 为socket写操作的hook
// 流程为:netSession start write -> PluginBeforeWrite... -> Write to socket
type PluginBeforeWrite interface {
	// BeforeWrite在真正执行socket读操作写调用
	// 通过添加插件,调用方可以改变即将使用的coder,meta,甚至改变write(不推荐)
	BeforeWrite(io.Writer, codec.Coder, interface{}) (io.Writer, codec.Coder, interface{})
}

// BeforeReadFunc 是PluginBeforeRead的Adaptor
type BeforeReadFunc func(io.Reader, codec.Coder, *message_meta.MessageMeta) (io.Reader, codec.Coder, *message_meta.MessageMeta)

// BeforeRead 在真正执行socket读操作前调用
// 通过添加插件,调用方可以改变即将使用的coder,meta,甚至改变reader(不推荐)
func (f BeforeReadFunc) BeforeRead(rdIn io.Reader, codecIn codec.Coder, metaIn *message_meta.MessageMeta) (io.Reader, codec.Coder, *message_meta.MessageMeta) {
	return f(rdIn, codecIn, metaIn)
}

// BeforeWriteFunc 是PluginBeforeRead的Adaptor
type BeforeWriteFunc func(io.Writer, codec.Coder, interface{}) (io.Writer, codec.Coder, interface{})

// BeforeWrite 在真正执行socket读操作写调用
// 通过添加插件,调用方可以改变即将使用的coder,meta,甚至改变write(不推荐)
func (f BeforeWriteFunc) BeforeWrite(wdIn io.Writer, codecIn codec.Coder, msgIn interface{}) (io.Writer, codec.Coder, interface{}) {
	return f(wdIn, codecIn, msgIn)
}
