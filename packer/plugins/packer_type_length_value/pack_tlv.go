package packer_type_length_value

import (
	"encoding/binary"
	"github.com/pkg/errors"
	"io"

	"github.com/MaxnSter/gnet/packer"
	"github.com/MaxnSter/gnet/util"
)

var (
	_ packer.Packer = (*tlvPacker)(nil)
)

const (
	Name = "tlv"

	LengthBytes = 4
	TypeBytes   = 4
	MaxLength   = 1 << 23
)

// ------------|---------------|-------------
// |  Length   |   Type(msgId) |   value	|
// |    4      |       4       |   msg  	|
// ------------------------------------------
// |           |--------------body----------|

type tlvPacker struct {
}

func UnpackMsgId(body []byte) (msgId uint32, value []byte) {
	msgId = binary.BigEndian.Uint32(body)
	value = body[TypeBytes:]
	return
}

func PackMsgId(msgId uint32, value []byte) (body []byte) {
	body = make([]byte, len(value)+TypeBytes)
	binary.BigEndian.PutUint32(body, msgId)
	copy(body[TypeBytes:], value)
	return
}

func (p *tlvPacker) Unpack(reader io.Reader) (body []byte, err error) {

	//读取长度段
	header := make([]byte, LengthBytes)
	_, err = io.ReadFull(reader, header)
	if err != nil {
		if err == io.ErrUnexpectedEOF {
			err = io.EOF
		}

		return
	}

	//解析长度段
	length := binary.BigEndian.Uint32(header)
	if length > MaxLength {
		err = errors.Errorf("msg too long, max:%d, actual:%d", MaxLength, length)
	}
	if length <= TypeBytes {
		err = errors.Errorf("msg too short, min:%d, actual:%d", TypeBytes, length)
	}
	if err != nil {
		return
	}

	//根据length,读取对应字节数
	body = make([]byte, length)
	//从body中解析messageId,根据messageId,我们可以获取该messageId对应的meta信息
	_, err = io.ReadFull(reader, body)

	return
}

func (p *tlvPacker) Pack(writer io.Writer, body []byte) error {
	if len(body) <= TypeBytes {
		return errors.Errorf("msg too short, min:%d, actual:%d", TypeBytes, len(body))
	}

	//对应上图, Length + Type + Value 总的长度
	totalLen := LengthBytes + len(body)

	//对应上图 Type + Value总的长度,
	bodyLen := len(body)

	pack := make([]byte, totalLen)

	// put length
	// Length的值 = Type + Value总的长度
	binary.BigEndian.PutUint32(pack, uint32(bodyLen))

	// put value([]byte after encode)
	copy(pack[LengthBytes:], body)

	return util.WriteFull(writer, pack)
}

// TypeName返回tlvPacker的名称
func (p *tlvPacker) String() string {
	return Name
}

func init() {
	packer.RegisterPacker(Name, &tlvPacker{})
}

func New() packer.Packer {
	return &tlvPacker{}
}
