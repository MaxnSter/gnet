package main

import (
	"fmt"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/example/echo"
	"github.com/MaxnSter/gnet/iface"
)

func main() {
	s := gnet.NewServer("127.0.0.1:9000", "server", func(ev iface.Event) {
		switch msg := ev.Message().(type) {
		case *echo.DemoMessage:
			fmt.Printf("msg receive:%s\n", msg.Val)

			ev.Session().Send(msg)
		}
	})

	s.StartAndRun()
}
