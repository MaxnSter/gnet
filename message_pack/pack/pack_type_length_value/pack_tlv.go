package pack_type_length_value

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/MaxnSter/gnet/codec"
	"github.com/MaxnSter/gnet/message_pack"
	"github.com/MaxnSter/gnet/message_pack/message_meta"
	"github.com/MaxnSter/gnet/util"
)

var (
	_ message_pack.Packer = (*tlvPacker)(nil)
)

const (
	tlvPackerName = "tlv"

	lengthBytes = 4
	typeBytes   = 4
	maxLength   = 1 << 23
)

// ------------|---------------|-------------
// |  Length   |   Type(msgId) |   value	|
// |    4      |       4       |   msg  	|
// ------------------------------------------
// |           |--------------body----------|

type tlvPacker struct {
}

// Unpack从指定reader中读数据,并根据指定的coder反序列化出一个message
func (p *tlvPacker) Unpack(reader io.Reader, c codec.Coder, meta *message_meta.MessageMeta) (msg interface{}, err error) {

	//读取长度段
	lengthBuf := make([]byte, lengthBytes)
	_, err = io.ReadFull(reader, lengthBuf)
	if err != nil {
		//remote close socket
		if err == io.ErrUnexpectedEOF {
			return nil, io.EOF
		}

		return nil, err
	}

	//解析长度段
	length := binary.BigEndian.Uint32(lengthBuf)
	if length > maxLength {
		return nil, errors.New("msg too long")
	}

	//根据length,读取对应字节数
	body := make([]byte, length)
	_, err = io.ReadFull(reader, body)
	if err != nil {
		return nil, err
	}

	//FIXME 优先使用指定的meta
	if meta != nil {
	}

	//从body中解析messageId,根据messageId,我们可以获取该messageId对应的meta信息
	msgID := binary.BigEndian.Uint32(body)
	msgNew := message_meta.MustGetMsgMeta(msgID).NewType()

	//body字段,meta信息,用于decode,得到最终的message
	body = body[typeBytes:]
	err = c.Decode(body, msgNew)
	if err != nil {
		return nil, err
	}

	return msgNew, nil
}

// Pack根据制定coder序列化制定message,最后写入制定writer
func (p *tlvPacker) Pack(writer io.Writer, c codec.Coder, msg interface{}) error {

	//获取该对应的messageId
	msgID := msg.(message_meta.MetaIdentifier).GetId()

	//对应上图中的value
	var buf []byte
	buf, err := c.Encode(msg)

	if err != nil {
		return err
	}

	//对应上图, Length + Type + Value 总的长度
	totalLen := lengthBytes + typeBytes + len(buf)

	//对应上图 Type + Value总的长度,
	bodyLen := typeBytes + len(buf)

	pack := make([]byte, totalLen)

	// put length
	// Length的值 = Type + Value总的长度
	binary.BigEndian.PutUint32(pack, uint32(bodyLen))

	// put type(msgID)
	binary.BigEndian.PutUint32(pack[lengthBytes:], uint32(msgID))

	// put value([]byte after encode)
	copy(pack[(lengthBytes+typeBytes):], buf)

	// 一直写,调用writeFull的原因见pack_type_length_value
	if err = util.WriteFull(writer, pack); err != nil {
		return err
	}

	return nil
}

// TypeName返回tlvPacker的名称
func (p *tlvPacker) TypeName() string {
	return tlvPackerName
}

func init() {
	message_pack.RegisterPacker(tlvPackerName, &tlvPacker{})
}
