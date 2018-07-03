package main

import (
	"bufio"
	"os"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/gnet/util"

	// 注册coder组件
	_ "github.com/MaxnSter/gnet/codec/codec_byte"

	// 注册packer组件
	_ "github.com/MaxnSter/gnet/message_pack/pack/pack_line"

	// 注册网络组件
	_ "github.com/MaxnSter/gnet/net/tcp"

	// 注册goroutine pool组件
	_ "github.com/MaxnSter/gnet/worker_pool/worker_session_race_other"
)

func loop(session gnet.NetSession) {
	scan := bufio.NewScanner(os.Stdin)
	for scan.Scan() {
		session.Send(scan.Text())
	}
	session.Stop()
}

func main() {
	module := gnet.NewModule(gnet.WithPool("poolRaceOther"), gnet.WithCoder("byte"),
		gnet.WithPacker("line"))
	operator := gnet.NewOperator(func(ev gnet.Event) {
		msg := ev.Message().([]byte)
		logger.Infoln("recv:", util.BytesToString(msg))
	})
	operator.SetOnConnected(func(session gnet.NetSession) {
		go loop(session)
	})

	client := gnet.NewNetClient("tcp", "echo", module, operator)
	client.Connect(":2007")
}
