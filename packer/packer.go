package packer

import (
	"io"
)

// Packer 定义了一个封包解包器
type Packer interface {
	Unpack(io.Reader) ([]byte, error)
	Pack(io.Writer, []byte) error
	String() string
}
