package main

import (
	"github.com/MaxnSter/gnet/example/memcached/memcached_server"
)

func main() {
	memcached_server.StartAndRun(":2007")
}
