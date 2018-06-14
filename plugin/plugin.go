// plugin allow us to control net readEvent and writeEvent
package plugin

import (
	"io"

	"github.com/MaxnSter/gnet/codec"
	"github.com/MaxnSter/gnet/message_pack/message_meta"
)

// PluginBeforeRead is hook before read from socket
// netSession start read -> PluginBeforeRead... -> read from socket
type PluginBeforeRead interface {
	// BeforeRead is call before read from socket
	// with this hook, you can change the reader, coder, and meta
	BeforeRead(io.Reader, codec.Coder, *message_meta.MessageMeta) (io.Reader, codec.Coder, *message_meta.MessageMeta)
}

// PluginBeforeWrite is hook before write to socket
// netSession start write -> PluginBeforeWrite... -> Write to socket
type PluginBeforeWrite interface {
	// BeforeWrite is call before write to socket
	// with this hook, you can change the write, coder, and msg
	BeforeWrite(io.Writer, codec.Coder, interface{}) (io.Writer, codec.Coder, interface{})
}

type BeforeReadFunc func(io.Reader, codec.Coder, *message_meta.MessageMeta) (io.Reader, codec.Coder, *message_meta.MessageMeta)

func (f BeforeReadFunc) BeforeRead(rdIn io.Reader, codecIn codec.Coder, metaIn *message_meta.MessageMeta) (io.Reader, codec.Coder, *message_meta.MessageMeta) {
	return f(rdIn, codecIn, metaIn)
}

type BeforeWriteFunc func(io.Writer, codec.Coder, interface{}) (io.Writer, codec.Coder, interface{})

func (f BeforeWriteFunc) BeforeWrite(wdIn io.Writer, codecIn codec.Coder, msgIn interface{}) (io.Writer, codec.Coder, interface{}) {
	return f(wdIn, codecIn, msgIn)
}
