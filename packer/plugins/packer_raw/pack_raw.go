package packer_raw

import (
	"github.com/MaxnSter/gnet/packer"
	"github.com/MaxnSter/gnet/util"
	"io"
)

const (
	BufSize = 1 << 10 * 8
)

type rawPacker struct {
}

func (p *rawPacker) Unpack(reader io.Reader) (buf []byte, err error) {
	var nRead int

	buf = make([]byte, BufSize)
	nRead, err = reader.Read(buf)
	buf = buf[:nRead]

	if err != nil {
		return
	}

	return
}

// Pack使用指定的coder序列化msg然后封包,最后写入writer
func (p *rawPacker) Pack(writer io.Writer, buf []byte) error {
	return util.WriteFull(writer, buf)
}

func (p *rawPacker) String() string {
	return "raw"
}

func init() {
	packer.RegisterPacker("raw", &rawPacker{})
}

func New() packer.Packer {
	return &rawPacker{}
}
