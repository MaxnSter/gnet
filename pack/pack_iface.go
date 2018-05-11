package pack

import (
	"io"

	"gnet/codec"
	"gnet/message"
)

type Packer interface {

	Unpack(reader io.Reader, c codec.Coder) (msg message.Message, err error)

	Pack(writer io.Writer, c codec.Coder, msg message.Message) error

	TypeName() string
}
