package pack_type_length_value

import (
	"encoding/binary"
	"errors"
	"io"

	"gnet/codec"
	"gnet/message"
	"gnet/pack"
)

var (
	_ pack.Packer = (*tlvPacker)(nil)
)

const (
	TlvPackerName = "tlv"
	LengthBytes   = 4
	TypeBytes     = 4

	// 8M
	MaxLength = 1 << 23
)

// ------------|---------------|-------------
// |  Length   |   Type(msgId) |   value	|
// |    4      |       4       |   bodyLen	|
// ------------------------------------------

type tlvPacker struct {
}

func (p *tlvPacker) Unpack(reader io.Reader, c codec.Coder) (msg message.Message, err error) {

	//read length of the Length
	lengthBuf := make([]byte, LengthBytes)
	_, err = io.ReadFull(reader, lengthBuf)
	if err != nil {
		return nil, err
	}

	//read the Length from bytes to int
	length := binary.LittleEndian.Uint32(lengthBuf)
	if length > MaxLength {
		return nil, errors.New("msg too big")
	}

	//read body
	body := make([]byte, length)
	_, err = io.ReadFull(reader, body)
	if err != nil {
		return nil, err
	}

	//get Type(msgId) and new a message
	msgId := binary.LittleEndian.Uint32(body)
	msgNew := message.MustGetMsgMeta(msgId).NewType()

	//decode
	body = body[TypeBytes:]
	err = c.Decode(body, msgNew)
	if err != nil {
		return nil, err
	}

	return msgNew.(message.Message), nil
}

func (p *tlvPacker) Pack(writer io.Writer, c codec.Coder, msg message.Message) error {
	msgId := msg.ID()

	var buf []byte
	buf, err := c.Encode(msg)

	if err != nil {
		return err
	}

	totalLen := LengthBytes + TypeBytes +  len(buf)
	bodyLen := TypeBytes + len(buf)
	pack := make([]byte, totalLen)

	//put length
	binary.LittleEndian.PutUint32(pack, uint32(bodyLen))

	//put type(msgId)
	binary.LittleEndian.PutUint32(pack[LengthBytes:], uint32(msgId))

	//put value([]byte after encode)
	copy(pack[(LengthBytes+TypeBytes):], buf)

	//write to writer
	//TODO must write fully
	if _, err := writer.Write(pack); err != nil {
		return err
	}

	return nil
}

func (p *tlvPacker) TypeName() string {
	return TlvPackerName
}

func init() {
	pack.RegisterPacker(TlvPackerName, &tlvPacker{})
}
