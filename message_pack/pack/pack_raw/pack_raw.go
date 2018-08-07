package pack_raw

import (
	"io"
	"github.com/MaxnSter/gnet/codec"
	"github.com/MaxnSter/gnet/message_pack/message_meta"
	"github.com/MaxnSter/gnet/message_pack"
	"github.com/MaxnSter/gnet/util"
)

const (
	bufSize = 1 << 10 * 8
)

type rawPacker struct {

}

func (p *rawPacker) Unpack(reader io.Reader, c codec.Coder, meta *message_meta.MessageMeta) (msg interface{}, err error) {
	var nRead int

	buf := make([]byte, bufSize)
	nRead, err = reader.Read(buf)
	buf = buf[:nRead]

	if err != nil {
		return nil, err
	}

	if meta != nil {
		newType := meta.NewType()
		err = c.Decode(buf, newType)
		msg = newType
	} else {
		var newType []byte
		err = c.Decode(buf, &newType)
		msg = newType
	}

	return
}

// Pack使用指定的coder序列化msg然后封包,最后写入writer
func (p *rawPacker) Pack(writer io.Writer, c codec.Coder, msg interface{}) error {
	data, err := c.Encode(msg)
	if err != nil {
		return nil
	}

	return util.WriteFull(writer, data)
}

func (p *rawPacker) TypeName() string {
	return "raw"
}

func init() {
	message_pack.RegisterPacker("raw", &rawPacker{})
}
