package main

import (
	"fmt"

	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/message/protocol/message_demo"
	"github.com/MaxnSter/gnet/net"
)

func main() {

	client := iface.NewClient("127.0.0.1:9000", func(ev net.Event) {
		switch msg := ev.Message().(type) {
		case *message_demo.DemoMessage:
			fmt.Printf("recv:%s\n", msg.Val)
		}
	}, iface.WithConnectedCB(func(session *net.TcpSession) {
		fmt.Println("connected")
		session.Send(message_demo.NewDemoMessage("123"))
	}))

	if err := client.Start(); err != nil {
		panic(err)
	}

	client.Run()
}
