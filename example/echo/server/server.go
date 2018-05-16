package main

import (
	"fmt"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/example/echo"
	"github.com/MaxnSter/gnet/iface"

	_ "github.com/MaxnSter/gnet/codec/codec_protobuf"
)

func main() {
	s := gnet.NewServer("127.0.0.1:9000", "server", func(ev iface.Event) {
		switch msg := ev.Message().(type) {
		case *echo.EchoProto:
			fmt.Printf("msg receive:%s\n", msg.Msg)

			ev.Session().Send(msg)
		}
	}, gnet.WithCoder("protoBuf"))

	s.StartAndRun()
}
