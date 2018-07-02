package main

import (
	"github.com/MaxnSter/gnet"

	// 注册coder组件
	_ "github.com/MaxnSter/gnet/codec/codec_byte"

	// 注册packer组件
	_ "github.com/MaxnSter/gnet/message_pack/pack/pack_line"

	// 注册网络组件
	_ "github.com/MaxnSter/gnet/net/tcp"

	// 注册goroutine pool组件
	_ "github.com/MaxnSter/gnet/worker_pool/worker_session_race_other"
)

// 业务逻辑入口
func onMessage(ev gnet.Event) {
	ev.Session().Send(ev.Message())
}

// telnet 127.0.0.1 2007
// nc 127.0.0.1 2007
func main() {
	// 指定封包组件packer,本处使用文本协议
	// 指定本次编解码组件coder
	// 指定goroutine池pool
	// 最后,我们创建一个了module对象
	module := gnet.NewModule(gnet.WithPacker("line"), gnet.WithCoder("byte"),
		gnet.WithPool("sessionRaceOther"))

	// 传入业务逻辑回调,创建一个module控制器
	operator := gnet.NewOperator(onMessage)

	// 指定我们的网络组件,传入module和module控制器,我们创建了一个NetServer
	server := gnet.NewNetServer("tcp", "echo", module, operator)

	// 启动服务器,可以用nc,telnet,或client测试
	server.ListenAndServe(":2007")
}
