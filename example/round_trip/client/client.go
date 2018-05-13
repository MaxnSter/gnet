package main

import (
	"fmt"
	"time"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/example/round_trip"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/message/protocol"
	"github.com/MaxnSter/gnet/net"
)

func getRoundTrip(tClient int64, tServer int64) {
	now := time.Now().UnixNano()
	avg := (now + tClient) / 2
	diff := tServer - avg

	fmt.Printf("round trip : %d, clock err : %d\n", now-tClient, diff)
}

func main() {
	client := gnet.NewClient("127.0.0.1:2007",
		func(ev iface.Event) {
			switch msg := ev.Message().(type) {
			case *round_trip.RoundTripProto:
				getRoundTrip(msg.T1, msg.T2)

				time.AfterFunc(time.Second, func() {

					msg := &round_trip.RoundTripProto{Id: protocol.ProtoRoundTrip, T1: time.Now().UnixNano(), T2: 0}

					ev.Session().Send(msg)
				})
			}
		},
		gnet.WithConnectedCB(func(s *net.TcpSession) {
			msg := &round_trip.RoundTripProto{Id: protocol.ProtoRoundTrip, T1: time.Now().UnixNano(), T2: 0}

			s.Send(msg)
		}))

	client.StartAndRun()
}
