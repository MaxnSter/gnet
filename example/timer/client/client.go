package main

import (
	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/example/timer"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/logger"

	_ "github.com/MaxnSter/gnet/codec/codec_msgpack"
)

func main() {

	gnetOption := gnet.NewGnetOption(gnet.WithCoder("msgspack"))
	client := gnet.NewClient("localhost:2007", nil, gnetOption, func(ev iface.Event) {
		switch msg := ev.Message().(type) {
		case *timer.TimerProto:
			logger.Debugf("receive msg, time:", msg.TimeNow.Format("Mon Jan 2 15:04:05 2006"))
		}
	})

	client.StartAndRun()
}
