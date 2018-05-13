package iface

import "io"

type Packer interface {
	Unpack(reader io.Reader, c Coder) (msg Message, err error)

	Pack(writer io.Writer, c Coder, msg Message) error

	TypeName() string
}
