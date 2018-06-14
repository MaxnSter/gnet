package main

import (
	"io"
	"reflect"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/codec"
	"github.com/MaxnSter/gnet/example/time"
	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/gnet/message_pack/message_meta"
	"github.com/MaxnSter/gnet/plugin"

	_ "github.com/MaxnSter/gnet/codec/codec_msgpack"
	_ "github.com/MaxnSter/gnet/message_pack/pack/pack_length_value"
	_ "github.com/MaxnSter/gnet/net/tcp"
	_ "github.com/MaxnSter/gnet/worker_pool/worker_session_race_other"
)

var (
	meta = message_meta.NewMessageMeta(1, reflect.TypeOf((*time.TimeProto)(nil)).Elem())
)

func main() {
	// 一般来说,客户端使用单eventLoop已经足够
	module := gnet.NewModule(gnet.WithCoder("msgpack"), gnet.WithPacker("lv"),
		gnet.WithPool("poolRaceOther"))
	module.SetRdPlugin(plugin.BeforeReadFunc(BeforeRead))

	o := gnet.NewOperator(onMessage)

	client := gnet.NewNetClient("tcp", "time", module, o)
	client.Connect(":2007")
}

// 未使用tlv packer情况下,可以通过plugin来自定义meta
func BeforeRead(rdIn io.Reader, codeIn codec.Coder, metaIn *message_meta.MessageMeta) (io.Reader, codec.Coder, *message_meta.MessageMeta) {
	return rdIn, codeIn, meta
}

func onMessage(ev gnet.Event) {
	switch msg := ev.Message().(type) {
	case *time.TimeProto:
		logger.Infoln("recv server key:", msg.Key, " server time:", msg.T.Unix())
	default:
		panic("error msg type")
	}
}
