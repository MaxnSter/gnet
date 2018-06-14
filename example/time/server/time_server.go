package main

import (
	"time"

	"github.com/MaxnSter/gnet"
	. "github.com/MaxnSter/gnet/example/time"
	"github.com/MaxnSter/gnet/iface"

	_ "github.com/MaxnSter/gnet/codec/codec_msgpack"
	_ "github.com/MaxnSter/gnet/message_pack/pack/pack_length_value"
	_ "github.com/MaxnSter/gnet/net/tcp"
	_ "github.com/MaxnSter/gnet/worker_pool/worker_session_norace"
)

func main() {
	// 无状态服务,可以使用poolNoRace,效率最高
	module := gnet.NewModule(gnet.WithPool("poolNoRace"), gnet.WithCoder("msgpack"),
		gnet.WithPacker("lv"))

	// 使用timer组件,timer组件需要一个pool来执行所有timeout回调
	module.SetTimer(module.Pool())

	o := gnet.NewOperator(func(ev gnet.Event) {})
	o.SetOnConnected(onConnected)

	server := gnet.NewNetServer("tcp", "time", module, o)
	server.ListenAndServe(":2007")
}

func onConnected(s gnet.NetSession) {
	// 回调中的ctx固定为NetSession
	s.RunEvery(time.Now(), time.Second, func(i time.Time, ctx iface.Context) {
		ctx.(gnet.NetSession).Send(TimeProto{T: time.Now(), Key:"i don't want to learn any more"})
	})
}
