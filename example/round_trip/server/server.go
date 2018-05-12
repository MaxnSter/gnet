package main

import (
	"time"

	"github.com/MaxnSter/gnet/example/round_trip"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/net"
)

func main() {
	server := iface.NewServer("127.0.0.1:2007", "", func(ev net.Event) {
		switch msg := ev.Message().(type) {
		case *round_trip.RoundTripProto:
			msg.T2 = time.Now().UnixNano()
			ev.Session().Send(msg)
		}
	})

	server.StartAndRun()
}
