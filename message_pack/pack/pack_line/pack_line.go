package pack_line

import (
	"bufio"
	"io"

	"github.com/MaxnSter/gnet/codec"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/message_pack"
	"github.com/MaxnSter/gnet/message_pack/message_meta"
	"github.com/MaxnSter/gnet/util"
)

const (
	// the name of linePacker
	linePackerName = "line"

	// message end mode
	lineEndModeLRLF = iota
	lineEndModeLF
)

var (
	_ message_pack.Packer = (*linePacker)(nil)

	lineEndModeKey lineEndModeKeyType = "lineEndModeKey"
)

type lineEndModeKeyType string

// linePack使用line Protocol
// 接收消息时解析并切割\r\n或\n结尾的消息
// 发送消息时自动以\r\n或\n结尾
type linePacker struct {
}

// Unpack read every single line end with \r\n from socket
func (p *linePacker) Unpack(reader io.Reader, c codec.Coder, meta *message_meta.MessageMeta) (msg interface{}, err error) {
	//FIXME,有点过于依赖api了
	rd, _ := reader.(*bufio.ReadWriter)
	buf, err := rd.ReadBytes('\n')

	// 如果reader能保存上下文,则保存当前消息是netcat模式(\n结尾)还是telnet模式(\r\n结尾)
	var ctxSetter iface.Property
	if c, ok := reader.(iface.Property); !ok {
		ctxSetter = nil
	} else {
		ctxSetter = c
	}

	if err != nil {
		//FIXME just a test for webSocket
		return nil, err
	}

	if len(buf) < 1 || buf[len(buf)-2] != '\r' {
		if ctxSetter != nil {
			ctxSetter.StoreCtx(lineEndModeKeyType("key"), lineEndModeLF)
		}
		buf = buf[:len(buf)-1]
	} else {
		if ctxSetter != nil {
			ctxSetter.StoreCtx(lineEndModeKeyType("key"), lineEndModeLRLF)
		}
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

// Packer 序列化msg后为消息加上\r\n或\n结尾,并写入制定writer
func (p *linePacker) Pack(writer io.Writer, c codec.Coder, msg interface{}) error {
	b, err := c.Encode(msg)
	if err != nil {
		return err
	}

	var ctxLoader iface.Property
	if c, ok := writer.(iface.Property); !ok {
		ctxLoader = nil
	} else {
		ctxLoader = c
	}

	if ctxLoader == nil {
		b = append(b, "\r\n"...)
	} else if v, ok := ctxLoader.LoadCtx(lineEndModeKey); ok && v.(int) == lineEndModeLF {
		b = append(b, "\n"...)
	} else {
		b = append(b, "\r\n"...)
	}

	err = util.WriteFull(writer, b)
	if err != nil {
		return nil
	}

	return nil
}

// TypeName返回linePacker的名称
func (p *linePacker) TypeName() string {
	return linePackerName
}

//register packer
func init() {
	message_pack.RegisterPacker(linePackerName, &linePacker{})
}
