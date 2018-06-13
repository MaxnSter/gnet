package pack_line

import (
	"bufio"
	"errors"
	"io"

	"github.com/MaxnSter/gnet/codec"
	"github.com/MaxnSter/gnet/message_pack"
	"github.com/MaxnSter/gnet/message_pack/message_meta"
	"github.com/MaxnSter/gnet/util"
)

const (
	// the name of linePacker
	LinePackerName = "line"
)

var (
	_ message_pack.Packer = (*linePacker)(nil)
)

// linePacker is packer for text protocol: line end with '\r\n'
type linePacker struct {
}

// Unpack read every single line end with \r\n from socket
func (p *linePacker) Unpack(reader io.Reader, c codec.Coder, meta *message_meta.MessageMeta) (msg interface{}, err error) {

	//FIXME
	rd, _ := reader.(*bufio.ReadWriter)
	buf, err := rd.ReadBytes('\n')

	if err != nil {
		return nil, err
	}

	if len(buf) < 1 || buf[len(buf)-2] != '\r' {
		//FIXME error wraps
		return nil, errors.New("msg not end with \r\n")
	} else {
		buf = buf[:len(buf)-2]
	}

	if meta != nil {
		newMsg := meta.NewType()
		err = c.Decode(buf, newMsg)
		if err != nil {
			return nil, err
		}
		msg = newMsg
	} else {
		var data []byte
		err = c.Decode(buf, &data)
		if err != nil {
			return nil, err
		}
		msg = data
	}

	return

}

// Packer encode msg and end with '\r\n', then send to socket
func (p *linePacker) Pack(writer io.Writer, c codec.Coder, msg interface{}) error {
	b, err := c.Encode(msg)
	if err != nil {
		return err
	}

	b = append(b, "\r\n"...)
	err = util.WriteFull(writer, b)
	if err != nil {
		return nil
	}

	return nil
}

// name of linePacker
func (p *linePacker) TypeName() string {
	return LinePackerName
}

//register packer
func init() {
	message_pack.RegisterPacker(LinePackerName, &linePacker{})
}
