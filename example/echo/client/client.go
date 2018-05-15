package main

import (
	"fmt"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/example/echo"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/net"
)

func main() {

	client := gnet.NewClient("127.0.0.1:9000",
		func(ev iface.Event) {
			switch msg := ev.Message().(type) {
			case *echo.DemoMessage:
				fmt.Printf("recv:%s\n", msg.Val)
			}
		},
		gnet.WithConnectedCB(func(session *net.TcpSession) {
			fmt.Println("connected")
			session.Send(echo.NewDemoMessage("123"))
		}))

	client.StartAndRun()
}
