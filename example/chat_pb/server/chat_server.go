package main

import (
	"net"

	// 注册使用的组件
	_ "github.com/MaxnSter/gnet/codec/codec_protobuf"
	_ "github.com/MaxnSter/gnet/message_pack/pack/pack_type_length_value"
	_ "github.com/MaxnSter/gnet/net/tcp"
	_ "github.com/MaxnSter/gnet/worker_pool/worker_session_race_self"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/example/chat_pb"
)

func main() {
	m := gnet.NewModule(gnet.WithPacker("tlv"),
		gnet.WithCoder("protobuf"),
		gnet.WithPool("poolRaceSelf"))

	o := gnet.NewOperator(onMessage)
	o.SetOnConnected(onConnected)

	server := gnet.NewNetServer("tcp", "chat_pb", m, o)
	server.ListenAndServe(":2007")
}

func onMessage(ev gnet.Event) {
	switch msg := ev.Message().(type) {
	case *chat_pb.ChatMessage:
		ev.Session().AccessManager().Broadcast(func(session gnet.NetSession) {
			session.Send(msg)
		})
	default:
		panic("error msg type")
	}
}

func onConnected(session gnet.NetSession) {
	msg := &chat_pb.ChatMessage{
		Id:     chat_pb.ChatMsgId,
		Talker: session.Raw().(net.Conn).LocalAddr().String(),
		Msg:    "welcome",
	}

	session.AccessManager().Broadcast(func(s gnet.NetSession) {
		s.Send(msg)
	})
}
