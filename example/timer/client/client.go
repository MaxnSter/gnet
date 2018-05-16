package main

import (
	"fmt"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/example/timer"
	"github.com/MaxnSter/gnet/iface"

	_ "github.com/MaxnSter/gnet/codec/codec_msgpack"
)

func main() {
	client := gnet.NewClient("127.0.0.1:9000", func(ev iface.Event) {
		switch msg := ev.Message().(type) {
		case *timer.TimerProto:
			fmt.Println("time:", msg.TimeNow)
		}
	}, gnet.WithCoder("msgpack"))

	client.StartAndRun()
}
