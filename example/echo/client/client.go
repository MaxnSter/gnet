package main

import (
	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/gnet/util"

	_ "github.com/MaxnSter/gnet/codec/codec_byte"
	_ "github.com/MaxnSter/gnet/pack/pack_line"
)

func main() {

	gnet.NewClient("127.0.0.1:2007", gnet.NewCallBackOption(), gnet.NewGnetOption(),
		func(ev iface.Event) {
			switch msg := ev.Message().(type) {
			case []byte:
				logger.WithField("msg", util.BytesToString(msg)).Debugln()
			}
		},
	).StartAndRun()
}
