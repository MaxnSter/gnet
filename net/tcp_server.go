package net

import (
	"net"
	"sync"
)

type TcpServer struct {
	options  NetOptions
	addr     string
	name 	 string
	listener net.Listener

	sessions sync.Map
}

func NewTcpServer(addr, name string, options NetOptions) *TcpServer {
	return nil
}
