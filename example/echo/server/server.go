package main

import (
	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/iface"

	_ "github.com/MaxnSter/gnet/codec/codec_byte"
	_ "github.com/MaxnSter/gnet/pack/pack_line"
)

func main() {
	gnetOption := gnet.NewGnetOption(gnet.WithCoder("byte"), gnet.WithPacker("line"))
	callback := gnet.NewCallBackOption()

	s := gnet.NewServer("0.0.0.0:2007", callback, gnetOption, func(ev iface.Event) {
		switch msg := ev.Message().(type) {
		case []byte, *[]byte:
			ev.Session().Send(msg)
		}
	})

	s.StartAndRun()
}
