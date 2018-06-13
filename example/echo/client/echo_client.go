package main

import (
	"bufio"
	"os"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/logger"
	_ "github.com/MaxnSter/gnet/net/tcp"
	"github.com/MaxnSter/gnet/util"
)

func loop(session gnet.NetSession) {
	scan := bufio.NewScanner(os.Stdin)
	for scan.Scan() {
		session.Send(scan.Text())
	}
}

func main() {
	module := gnet.NewDefaultModule()
	operator := gnet.NewOperator(func(ev gnet.Event) {
		switch msg := ev.Message().(type) {
		case []byte:
			logger.Infoln("recv:" , util.BytesToString(msg))
		}
	})
	operator.SetOnConnected(func(session gnet.NetSession) {
		go loop(session)
	})

	client := gnet.NewNetClient("tcp", "echo", module, operator)
	client.Connect(":2007")
}
