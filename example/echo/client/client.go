package main

import (
	"bufio"
	"os"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/gnet/net"
	"github.com/MaxnSter/gnet/util"

	_ "github.com/MaxnSter/gnet/codec/codec_byte"
	_ "github.com/MaxnSter/gnet/pack/pack_text"
)

func readLoop(s *net.TcpSession) {

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		s.Send(scanner.Bytes())
	}

}

func main() {

	gnet.NewClient("127.0.0.1:2007",
		func(ev iface.Event) {
			switch msg := ev.Message().(type) {
			case []byte:
				logger.WithField("msg", util.BytesToString(msg)).Debugln()
			}
		},

		gnet.WithConnectedCB(func(session *net.TcpSession) {
			go readLoop(session)
		}),

		gnet.WithCoder("byte"), gnet.WithPacker("text")).StartAndRun()
}
