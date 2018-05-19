package pack_type_length_value

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/message"
	"github.com/MaxnSter/gnet/pack"
	"github.com/MaxnSter/gnet/util"
)

var (
	_ iface.Packer = (*tlvPacker)(nil)
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
// |    4      |       4       |   msg  	|
// ------------------------------------------
// |           |--------------body----------|

type tlvPacker struct {
}

//从指定reader中读数据,并根据指定的coder反序列化出一个message
func (p *tlvPacker) Unpack(reader io.Reader, c iface.Coder) (msg iface.Message, err error) {

	//读取长度段
	lengthBuf := make([]byte, LengthBytes)
	_, err = io.ReadFull(reader, lengthBuf)
	if err != nil {
		//remote close socket
		if err == io.ErrUnexpectedEOF {
			return nil, io.EOF
		}

		return nil, err
	}

	//解析长度段
	length := binary.LittleEndian.Uint32(lengthBuf)
	if length > MaxLength {
		return nil, errors.New("msg too big")
	}

	//根据length,读取对应字节数
	body := make([]byte, length)
	_, err = io.ReadFull(reader, body)
	if err != nil {
		return nil, err
	}

	//从body中解析messageId,根据messageId,我们可以获取该messageId对应的meta信息
	msgId := binary.LittleEndian.Uint32(body)
	msgNew := message.MustGetMsgMeta(msgId).NewType()

	//body字段,meta信息,用于decode,得到最终的message
	body = body[TypeBytes:]
	err = c.Decode(body, msgNew)
	if err != nil {
		return nil, err
	}

	return msgNew.(iface.Message), nil
}

//根据制定coder序列化制定message,最后写入制定writer
func (p *tlvPacker) Pack(writer io.Writer, c iface.Coder, msg iface.Message) error {

	//获取该对应的messageId
	msgId := msg.GetId()

	//对应上图中的value
	var buf []byte
	buf, err := c.Encode(msg)

	if err != nil {
		return err
	}

	//对应上图, Length + Type + Value 总的长度
	totalLen := LengthBytes + TypeBytes + len(buf)

	//对应上图 Type + Value总的长度,
	bodyLen := TypeBytes + len(buf)

	pack := make([]byte, totalLen)

	// put length
	// Length的值 = Type + Value总的长度
	binary.LittleEndian.PutUint32(pack, uint32(bodyLen))

	// put type(msgId)
	binary.LittleEndian.PutUint32(pack[LengthBytes:], uint32(msgId))

	// put value([]byte after encode)
	copy(pack[(LengthBytes+TypeBytes):], buf)

	// 一直写
	if err := util.WriteFull(writer, pack); err != nil {
		return err
	}

	return nil
}

//name of tlvPacker
func (p *tlvPacker) TypeName() string {
	return TlvPackerName
}

//注册packer
func init() {
	pack.RegisterPacker(TlvPackerName, &tlvPacker{})
}
