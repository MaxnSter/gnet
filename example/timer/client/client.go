package main

import (
	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/example/timer"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/logger"

	_ "github.com/MaxnSter/gnet/codec/codec_msgpack"
)

func main() {
	client := gnet.NewClient("127.0.0.1:2007", func(ev iface.Event) {
		switch msg := ev.Message().(type) {
		case *timer.TimerProto:
			logger.Debugf("receive msg, time:", msg.TimeNow.Format("Mon Jan 2 15:04:05 2006"))
		}
	}, gnet.WithCoder("msgpack"))

	client.StartAndRun()
}
