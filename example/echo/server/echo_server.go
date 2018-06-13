package main

import (
	"github.com/MaxnSter/gnet"
	_ "github.com/MaxnSter/gnet/net/tcp"
)

func onMessage(ev gnet.Event) {
	ev.Session().Send(ev.Message())
}

func main() {
	module := gnet.NewDefaultModule()
	operator := gnet.NewOperator(onMessage)

	server := gnet.NewNetServer("tcp", "echo", module, operator)
	server.ListenAndServe(":2007")
}
