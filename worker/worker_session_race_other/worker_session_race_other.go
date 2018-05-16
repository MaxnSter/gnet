package worker_session_race_other

import (
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/worker"
	"github.com/MaxnSter/gnet/worker/internal/basic_event_queue"
)

const (
	poolName  = "poolRaceOther"
	queueSize = 500
)

func init() {
	worker.RegisterWorkerPool(poolName, NewPoolRaceOther)
}

//session存在data race,应且频繁与其他session交互
//此时,一个worker负责所有session的事件处理,比如MMOARPG的逻辑线程
type poolRaceOther struct {
	//TODO resizeable channel?
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
		//TODO logs
		p.queue.MustPut(cb)
	}
}

func (p *poolRaceOther) TryPut(session iface.NetSession, cb func()) bool {

	if err := p.queue.Put(cb); err != nil {
		return false
	}

	return true
}
