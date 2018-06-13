package main

import (
	"bufio"
	"flag"
	"os"
	"strings"

	"github.com/MaxnSter/gnet/example/chat/client"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/gnet/net"
	"github.com/MaxnSter/gnet/util"
)

func main() {
	addr := flag.String("addr", ":2007", "server address")
	flag.Parse()

	client.NewChatClient(*addr, readConsole, recvMessage).Run()
}

func readConsole(session *net.TcpSession) {
	go func() {
		logger.Infoln("waiting for input...")
		scan := bufio.NewScanner(os.Stdin)
		for scan.Scan() {
			tx := scan.Text()

			if strings.ToUpper(tx) == "Q" {
				break
			} else {
				session.Send(tx)
			}
		}
		session.Stop()
	}()

}

func recvMessage(ev iface.Event) {
	logger.Infoln("recv:", util.BytesToString(ev.Message().([]byte)))
}
