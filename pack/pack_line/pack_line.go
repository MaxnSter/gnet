package pack_line

import (
	"bufio"
	"bytes"
	"errors"
	"io"

	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/pack"
	"github.com/MaxnSter/gnet/util"
)

const (
	// the name of linePacker
	LinePackerName = "line"
)

var (
	_ iface.Packer = (*linePacker)(nil)

	scannerEOF = errors.New("scanner eof")
)

// linePacker is packer for text protocol: line end with '\r\n'
// NOTE: Must use with byte coder
type linePacker struct {
}

// Unpack read every single line end with \r\n from socket
func (p *linePacker) Unpack(reader io.Reader, c iface.Coder) (msg interface{}, err error) {

	//TODO
	rd, _ := reader.(*bufio.ReadWriter)
	buf, err := rd.ReadBytes('\n')

	if err != nil {
		return nil, err
	}

	if len(buf) < 1 || buf[len(buf)-2] != '\r' {
		return nil, errors.New("msg not end with \r\n")
	} else {
		buf = buf[:len(buf)-2]
	}

	var data []byte
	err = c.Decode(buf, &data)

	if err != nil {
		return nil, err
	}

	msg = data
	return msg, nil

}

// Packer encode msg and end with '\r\n', then send to socket
func (p *linePacker) Pack(writer io.Writer, c iface.Coder, msg interface{}) error {
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

// split is func for scanner, check line is ends with \r\n
func (p *linePacker) split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, scannerEOF
	}

	if i := bytes.IndexByte(data, '\n'); i >= 1 {

		if data[i-1] != '\r' {
			return 0, nil, errors.New("msg not end with \r\n")
		}

		return i + 1, data[0 : i-1], nil
	}

	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return 0, nil, scannerEOF
	}
	// Request more data.
	return 0, nil, nil
}

//register packer
func init() {
	pack.RegisterPacker(LinePackerName, &linePacker{})
}
