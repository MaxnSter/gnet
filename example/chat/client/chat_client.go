package main

import (
	"bufio"
	"net"
	"os"

	// import使用的组件对应的包
	_ "github.com/MaxnSter/gnet/codec/codec_protobuf"
	_ "github.com/MaxnSter/gnet/message_pack/pack/pack_type_length_value"
	_ "github.com/MaxnSter/gnet/net/tcp"
	_ "github.com/MaxnSter/gnet/worker_pool/worker_session_race_other"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/example/chat"
	"github.com/MaxnSter/gnet/logger"
)

func main() {
	// 注册组件,使用protobuf编解码,type_length_value的封包解包方式,单个eventLoop处理所有非io事件
	// 客户端服务端使用coder, packer必须相同
	m := gnet.NewModule(gnet.WithCoder("protobuf"), gnet.WithPacker("tlv"),
		gnet.WithPool("poolRaceOther"))

	o := gnet.NewOperator(onMessage)
	o.SetOnConnected(func(session gnet.NetSession) {
		go loop(session)
	})

	client := gnet.NewNetClient("tcp", "chat", m, o)
	client.Connect(":2007")
}

func loop(s gnet.NetSession) {
	scan := bufio.NewScanner(os.Stdin)
	for scan.Scan() {
		msg := &chat.ChatMessage{
			Id:     chat.ChatMsgId,
			Talker: s.Raw().(net.Conn).RemoteAddr().String(),
			Msg:    scan.Text(),
		}
		s.Send(msg)
	}
	s.Stop()
}

func onMessage(ev gnet.Event) {
	switch msg := ev.Message().(type) {
	case *chat.ChatMessage:
		logger.Infoln("recv msg:" + msg.Msg + " from:" + msg.Talker)
	}
}
