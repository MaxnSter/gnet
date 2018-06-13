package message_pack

import (
	"io"

	"github.com/MaxnSter/gnet/codec"
	"github.com/MaxnSter/gnet/message_pack/message_meta"
)

type Packer interface {
	Unpack(reader io.Reader, c codec.Coder, meta *message_meta.MessageMeta) (msg interface{}, err error)

	Pack(writer io.Writer, c codec.Coder, msg interface{}) error

	TypeName() string
}
