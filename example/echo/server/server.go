package main

import (
	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/iface"

	_ "github.com/MaxnSter/gnet/codec/codec_byte"
	_ "github.com/MaxnSter/gnet/pack/pack_text"
)

func main() {
	s := gnet.NewServer("0.0.0.0:2007", "server", func(ev iface.Event) {
		switch msg := ev.Message().(type) {
		case []byte, *[]byte:
			ev.Session().Send(msg)
		}
	}, gnet.WithCoder("byte"), gnet.WithPacker("text"))

	s.StartAndRun()
}
