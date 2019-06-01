package pack_length_value

import (
	"encoding/binary"
	"github.com/pkg/errors"
	"io"

	"github.com/MaxnSter/gnet/packer"
	"github.com/MaxnSter/gnet/util"
)

const (
	// the name of pack_length_value
	Name = "lv"

	// LengthSize
	LengthSize = 4

	// MaxLen 8M
	MaxLen = 1 << 23
)

// ------------|---------------|-------------
// |  Length   |             value	        |
// |    4      |                            |
// ------------------------------------------

var (
	_ packer.Packer = (*lvPacker)(nil)
)

type lvPacker struct {
}

// Unpack 使用length-value的解包方式读消息,然后返回value对应的[]byte
func (p *lvPacker) Unpack(reader io.Reader) (value []byte, err error) {
	// 读取长度段
	lengthBuf := make([]byte, LengthSize)
	_, err = io.ReadFull(reader, lengthBuf)
	if err != nil {
		// readFull把io.EoF视为io.ErrUnexpectedEOF
		if err == io.ErrUnexpectedEOF {
			err = io.EOF
		}

		return
	}

	// 解析长度字段
	length := binary.BigEndian.Uint32(lengthBuf)
	if length > MaxLen {
		err = errors.Errorf("msg too long, max:%d, actual:%d", MaxLen, length)
		return
	}

	// 根据长度读取对应长度的字节
	value = make([]byte, length)
	_, err = io.ReadFull(reader, value)

	return
}

// Pack 使用length-value的形式对消息封包,并保证全部写入socket,直到错误
func (p *lvPacker) Pack(writer io.Writer, value []byte) error {
	// 对着上图,totalLen := lenSize + valueLen
	valueLen := len(value)
	totalLen := LengthSize + valueLen
	pack := make([]byte, totalLen)

	// 写入length段
	binary.BigEndian.PutUint32(pack, uint32(valueLen))

	// 写入value段
	copy(pack[LengthSize:], value)

	// 一直写,虽然runtime在write socket时也会保证全部写入(源码里有)
	// 但这里的writer对应的不一定是io.conn,也有可能是包装buffer之后的
	// writer,所以,还是需要调用writeFull滴
	if err := util.WriteFull(writer, pack); err != nil {
		return err
	}

	return nil
}

// TypeName 返回lvPacker的名称
func (p *lvPacker) String() string {
	return Name
}

func init() {
	packer.RegisterPacker(Name, &lvPacker{})
}

func New() packer.Packer {
	return &lvPacker{}
}
