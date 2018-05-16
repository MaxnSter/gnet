package main

import (
	"time"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/example"
	"github.com/MaxnSter/gnet/example/timer"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/net"

	_ "github.com/MaxnSter/gnet/codec/codec_msgpack"
)

func main() {
	onEvent := func(ev iface.Event) {}
	onTimer := func(t time.Time, s iface.NetSession) {
		s.Send(&timer.TimerProto{
			Id:      example.ProtoTimer,
			TimeNow: time.Now(),
		})
	}
	onConnect := func(s *net.TcpSession) {
		s.RunEvery(time.Now(), time.Second*5, onTimer)
	}

	server := gnet.NewServer("127.0.0.1:9000", "server", onEvent,
		gnet.WithConnectedCB(onConnect), gnet.WithCoder("msgpack"))
	server.StartAndRun()
}
