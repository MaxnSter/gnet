package main

import (
	"fmt"
	"net"

	"github.com/MaxnSter/gnet"

	//注册我们使用的组件
	_ "github.com/MaxnSter/gnet/codec/codec_byte"                      //注册编解码组件
	_ "github.com/MaxnSter/gnet/message_pack/pack/pack_line"           //注册封解包组件
	_ "github.com/MaxnSter/gnet/net/tcp"                               //注册网络组件
	_ "github.com/MaxnSter/gnet/worker_pool/worker_session_race_other" //注册并发池组件
)

// 加入聊天组
func onConnected(s gnet.NetSession) {
	msg := fmt.Sprintf("user:%s, join", s.Raw().(net.Conn).RemoteAddr())

	//广播有用户加入
	s.AccessManager().Broadcast(func(session gnet.NetSession) {
		session.Send(msg)
	})
}

// 离开聊天组
func onClose(s gnet.NetSession) {
	msg := fmt.Sprintf("user:%s, leave", s.Raw().(net.Conn).RemoteAddr())

	//广播有用户离开
	s.AccessManager().Broadcast(func(session gnet.NetSession) {
		session.Send(msg)
	})
}

// 接收聊天消息
func onMessage(ev gnet.Event) {
	s := ev.Session()
	msg := fmt.Sprintf("user:%s, talk:%s", s.Raw().(net.Conn).RemoteAddr(),
		ev.Message())

	// 广播聊天消息给其他用户
	s.AccessManager().Broadcast(func(session gnet.NetSession) {
		session.Send(msg)
	})
}

func main() {
	//使用裸包编解码方式,文本协议的封解包方式,单EventLoop的并发模型,创建一个module
	module := gnet.NewModule(gnet.WithCoder("byte"), gnet.WithPacker("line"),
		gnet.WithPool("poolRaceOther"))

	//创建module控制器,注册相关回调
	operator := gnet.NewOperator(onMessage)
	operator.SetOnConnected(onConnected)
	operator.SetOnClose(onClose)

	//使用tcp网络组件,传入module及其控制器,创建我们的gnet server
	//启动成功后,使用telnet或netcat测试
	server := gnet.NewNetServer("tcp", "chat", module, operator)
	server.ListenAndServe("127.0.0.1:8000")
}
