package worker_session_norace

import (
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/net"
	"github.com/MaxnSter/gnet/worker"
)

const (
	poolName = "poolNoRace"
)

func init()  {
	worker.RegisterWorkerPool(poolName, NewPoolNoRace)
}

type poolNoRace struct {

}

func (p *poolNoRace) TypeName() string {
	return poolName
}

func NewPoolNoRace() iface.WorkerPool{

}

func (p *poolNoRace) Start() {

}

func (p *poolNoRace) Stop() (done <- chan struct{}) {

}

func (p *poolNoRace) Put(session iface.NetSession, cb func())  {

}



