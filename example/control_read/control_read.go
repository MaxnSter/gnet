package main

import (
	"bufio"
	"io"
	"os"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/codec"
	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/gnet/message_pack/message_meta"
	_ "github.com/MaxnSter/gnet/net/tcp"
	"github.com/MaxnSter/gnet/plugin"
	"github.com/MaxnSter/gnet/util"
)

var (
	controlCh = make(chan struct{})
)

func release() {
	select {
	case <-controlCh:
		logger.Infoln("release session read")
	default:
	}
}

func block() {
	logger.Infoln("block session read")
	controlCh <- struct{}{}
}

func controlFromStdin(session gnet.NetSession) {
	scan := bufio.NewScanner(os.Stdin)
	for scan.Scan() {
		release()
	}
	session.Stop()
}

func onConnect(session gnet.NetSession) {
	go controlFromStdin(session)
}

func onMessage(ev gnet.Event) {
	logger.Infoln("recv:", util.BytesToString(ev.Message().([]byte)))
}

// telnet 127.0.0.1 2007
func main() {
	//增加plugin拦截读操作
	module := gnet.NewDefaultModule()
	module.SetRdPlugin(plugin.BeforeReadFunc(beforeRead))

	o := gnet.NewOperator(onMessage)
	o.SetOnConnected(onConnect)

	server := gnet.NewNetServer("tcp", "", module, o)
	server.ListenAndServe(":2007")
}

func beforeRead(rdIn io.Reader, codecIn codec.Coder, metaIn *message_meta.MessageMeta) (io.Reader, codec.Coder, *message_meta.MessageMeta) {
	block()
	return rdIn, codecIn, metaIn
}