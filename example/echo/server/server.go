package main

import (
	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/example/echo"
	"github.com/MaxnSter/gnet/iface"

	_ "github.com/MaxnSter/gnet/codec/codec_protobuf"
)

func main() {
	s := gnet.NewServer("0.0.0.0:2007", "server", func(ev iface.Event) {
		switch msg := ev.Message().(type) {
		case *echo.EchoProto:
			ev.Session().Send(msg)
		}
	}, gnet.WithCoder("protoBuf"))

	s.StartAndRun()
}
