package main

import (
	"math/rand"
	"time"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/example"
	"github.com/MaxnSter/gnet/example/timer"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/net"

	_ "github.com/MaxnSter/gnet/codec/codec_msgpack"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	onEvent := func(ev iface.Event) {}
	onTimer := func(t time.Time, ctx iface.Context) {
		s := ctx.(iface.NetSession)
		s.Send(&timer.TimerProto{
			Id:      example.ProtoTimer,
			TimeNow: time.Now(),
		})
	}
	onConnect := func(s *net.TcpSession) {
		r := rand.Intn(5) + 1
		s.RunEvery(time.Now(), time.Duration(r)*time.Second, onTimer)
		s.Send(&timer.TimerProto{
			Id:      example.ProtoTimer,
			TimeNow: time.Now(),
		})
	}

	callback := gnet.NewCallBackOption(gnet.WithOnConnectCB(onConnect))
	gnetOption := gnet.NewGnetOption(gnet.WithCoder("msgpack"))
	server := gnet.NewServer("0.0.0.0:2007", callback, gnetOption, onEvent)
	server.StartAndRun()
}
