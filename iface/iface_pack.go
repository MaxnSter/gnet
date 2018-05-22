package iface

import "io"

type Packer interface {
	Unpack(reader io.Reader, c Coder) (msg interface{}, err error)

	Pack(writer io.Writer, c Coder, msg interface{}) error

	TypeName() string
}
