package pack

import (
	"io"

	"github.com/MaxnSter/gnet/codec"
	"github.com/MaxnSter/gnet/message"
)

type Packer interface {
	Unpack(reader io.Reader, c codec.Coder) (msg message.Message, err error)

	Pack(writer io.Writer, c codec.Coder, msg message.Message) error

	TypeName() string
}
