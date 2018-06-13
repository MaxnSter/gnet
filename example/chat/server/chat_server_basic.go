package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"

	_ "github.com/MaxnSter/gnet/codec/codec_byte"
	_ "github.com/MaxnSter/gnet/message_pack/pack/pack_length_value"
	_ "github.com/MaxnSter/gnet/worker_pool/worker_session_race_other"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/net"
)

func main() {
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			panic(err)
		}
	}()
	port := flag.String("p", "2007", "listen port")
	flag.Parse()

	server := &BasicChatServer{}
	server.Start(*port)
}

type BasicChatServer struct {
	*net.TcpServer
}

func (s *BasicChatServer) Start(port string) {
	netOption := &gnet.GnetOption{Coder: "byte", Packer: "lv", WorkerPool: "poolRaceOther"}
	s.TcpServer = gnet.NewServer("0.0.0.0:"+port, &gnet.CallBackOption{}, netOption, s.onMessage)
	s.StartAndRun()
}

func (s *BasicChatServer) onMessage(ev iface.Event) {
	switch msg := ev.Message().(type) {
	case []byte:
		s.TcpServer.AccessSession(func(Id, session interface{}) bool {
			session.(iface.NetSession).Send(msg)
			return true
		})
	}
}
