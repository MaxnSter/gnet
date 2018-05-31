package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/MaxnSter/gnet/example/memcached/memcached_server"
)

func main() {

	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			panic(err)
		}
	}()
	memcached_server.StartAndRun(":2007")
}
