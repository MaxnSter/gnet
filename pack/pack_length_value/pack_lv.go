package pack_length_value

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/pack"
	"github.com/MaxnSter/gnet/util"
)

const (
	// the name of pack_length_value
	LvPackName = "lv"

	// lengthSize
	lengthSize = 4

	// maxLen 8M
	maxLen = 1 << 23
)

// ------------|---------------|-------------
// |  Length   |             value	        |
// |    4      |                            |
// ------------------------------------------

var (
	_ iface.Packer = (*lvPacker)(nil)
)

// NOTE: lvPacker只能与byteCoder使用
type lvPacker struct {
}

// Unpack 使用length-value的解包方式读消息,然后返回value对应的[]byte
func (p *lvPacker) Unpack(reader io.Reader, c iface.Coder) (msg interface{}, err error) {
	// 读取长度段
	lengthBuf := make([]byte, lengthSize)
	_, err = io.ReadFull(reader, lengthBuf)
	if err != nil {
		// readFull把io.EoF视为io.ErrUnexpectedEOF
		if err == io.ErrUnexpectedEOF {
			return nil, io.EOF
		}

		return nil, err
	}

	// 解析长度字段
	length := binary.BigEndian.Uint32(lengthBuf)
	if length > maxLen {
		//TODO error wrapper
		return nil, errors.New("msg too long")
	}

	// 根据长度读取对应长度的字节
	value := make([]byte, length)
	_, err = io.ReadFull(reader, value)
	if err != nil {
		return nil, err
	}

	// 不知道对应的meta,所以只能用byteCoder
	// 未来考虑这里可以指定一个meta
	var data[]byte
	err = c.Decode(value, &data)
	if err != nil {
		return nil, err
	}

	msg = data
	return
}

// Pack 使用length-value的形式对消息封包,并保证全部写入socket,直到错误
func (p *lvPacker) Pack(writer io.Writer, c iface.Coder, msg interface{}) error {
	encodeBuf, err := c.Encode(msg)
	if err != nil {
		return err
	}

	// 对着上图,totalLen := lenSize + valueLen
	valueLen := len(encodeBuf)
	totalLen := lengthSize + valueLen
	pack := make([]byte, totalLen)

	// 写入length段
	binary.BigEndian.PutUint32(pack, uint32(valueLen))

	// 写入value段
	copy(pack[lengthSize:], encodeBuf)

	// 一直写,虽然runtime在write socket时也会保证全部写入(源码里有)
	// 但这里的writer对应的不一定是io.conn,也有可能是包装buffer之后的
	// writer,所以,还是需要调用writeFull滴
	if err := util.WriteFull(writer, pack); err != nil {
		return err
	}

	return nil
}

// TypeName 返回lvPacker的名称
func (p *lvPacker) TypeName() string{
	return LvPackName
}

func init() {
	pack.RegisterPacker(LvPackName, &lvPacker{})
}
