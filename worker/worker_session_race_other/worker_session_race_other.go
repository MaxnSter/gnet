package worker_session_race_other

import (
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/worker"
	"github.com/MaxnSter/gnet/worker/internal/basic_event_queue"
)

const (
	poolName  = "poolRaceOther"
	queueSize = 100
)

func init() {
	worker.RegisterWorkerPool(poolName, NewPoolRaceOther)
}

type poolRaceOther struct {
	queue *basic_event_queue.EventQueue
}

func (p *poolRaceOther) TypeName() string {
	return poolName
}

func NewPoolRaceOther() iface.WorkerPool {
	return &poolRaceOther{
		queue: basic_event_queue.NewEventQueue(queueSize, true),
	}
}

func (p *poolRaceOther) Start() {
	p.queue.Start()
}

func (p *poolRaceOther) Stop() (done <-chan struct{}) {
	return p.queue.Stop()
}

func (p *poolRaceOther) Put(session iface.NetSession, cb func()) {
	if err := p.queue.Put(cb); err != nil {
		//TODO
		for p.queue.Put(cb) != nil {
		}
	}
}
