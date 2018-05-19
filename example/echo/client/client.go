package main

import (
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/example"
	"github.com/MaxnSter/gnet/example/echo"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/gnet/net"

	_ "github.com/MaxnSter/gnet/codec/codec_protobuf"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {

	clientNum := flag.Int("c", 1, "concurrency client number")
	flag.Parse()

	fmt.Println("//////////////////////////////////////////////")
	fmt.Println("//         concurrency number : ", *clientNum, "///////////")
	fmt.Println("//////////////////////////////////////////////")

	wg := sync.WaitGroup{}

	for i := 0; i < *clientNum; i++ {

		wg.Add(1)
		go func() {

			var sendTime int64

			gnet.NewClient("aliyun:2007",
				func(ev iface.Event) {
					switch ev.Message().(type) {
					case *echo.EchoProto:
						curTime := time.Now().UnixNano() / 1e6
						logger.WithField("ttl", curTime-sendTime).Debugln("received")

						t := time.Duration(rand.Intn(3) + 1)
						time.Sleep(t * time.Second)
						sendTime = time.Now().UnixNano() / 1e6
						ev.Session().Send(&echo.EchoProto{Id: example.ProtoEcho, Msg: "1234567890"})

					}
				},
				gnet.WithConnectedCB(func(session *net.TcpSession) {
					sendTime = time.Now().UnixNano() / 1e6
					session.Send(&echo.EchoProto{Id: example.ProtoEcho, Msg: "1234567890"})
				}), gnet.WithCoder("protoBuf")).StartAndRun()

			wg.Done()
		}()
	}

	wg.Wait()

}
