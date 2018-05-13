package main

import (
	"fmt"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/message/protocol/message_demo"
	"github.com/MaxnSter/gnet/net"
)

func main() {
	s := gnet.NewServer("127.0.0.1:9000", "server", func(ev iface.Event) {
		switch msg := ev.Message().(type) {
		case *message_demo.DemoMessage:
			fmt.Printf("msg receive:%s\n", msg.Val)

			ev.Session().Send(msg)
		}
	})

	s.StartAndRun()
}
