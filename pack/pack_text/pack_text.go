package pack_text

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
	// the name of textPacker
	TextPackerName = "text"
)

var (
	_ iface.Packer = (*textPacker)(nil)

	scannerEOF = errors.New("scanner eof")
)

// textPacker is packer for text protocol: line end with '\r\n'
// NOTE: Must use with byte coder
type textPacker struct {
}

// Unpack read every single line end with \r\n from socket
func (p *textPacker) Unpack(reader io.Reader, c iface.Coder) (msg interface{}, err error) {
	scan := bufio.NewScanner(reader)
	scan.Split(p.split)

	if scan.Scan() {
		var b []byte

		err := c.Decode(scan.Bytes(), &b)
		if err != nil {
			return nil, err
		}
		return b, nil

	} else {
		err = scan.Err()
		if err == scannerEOF {
			err = io.EOF
		}

		return nil, err
	}
}

// Packer encode msg and end with '\r\n', then send to socket
func (p *textPacker) Pack(writer io.Writer, c iface.Coder, msg interface{}) error {
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

// name of textPacker
func (p *textPacker) TypeName() string {
	return TextPackerName
}

// split is func for scanner, check line is ends with \r\n
func (p *textPacker) split(data []byte, atEOF bool) (advance int, token []byte, err error) {
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
	pack.RegisterPacker(TextPackerName, &textPacker{})
}
