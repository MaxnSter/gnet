package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/example"
	"github.com/MaxnSter/gnet/example/echo"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/net"

	_ "github.com/MaxnSter/gnet/codec/codec_protobuf"
)

func echoLoop(s iface.NetSession) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		s.Send(&echo.EchoProto{example.ProtoEcho, scanner.Text()})
	}
	s.Stop()
}

func main() {

	client := gnet.NewClient("127.0.0.1:9000",
		func(ev iface.Event) {
			switch msg := ev.Message().(type) {
			case *echo.EchoProto:
				fmt.Printf("recv:%s\n", msg.Msg)
			}
		},
		gnet.WithConnectedCB(func(session *net.TcpSession) {
			go echoLoop(session)
		}), gnet.WithCoder("protoBuf"))

	client.StartAndRun()
}
