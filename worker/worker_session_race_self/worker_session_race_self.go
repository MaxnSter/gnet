package worker_session_race_self

import (
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/worker"
	"github.com/MaxnSter/gnet/worker/internal/basic_event_queue"
)

const (
	poolName = "poolRaceSelf"

	workerNum = 20
	queueSize = 100
)

func init() {
	worker.RegisterWorkerPool(poolName, NewPoolRaceSelf)
}

//session存在data race现象,并且几乎没有其他session交互的情况
//此时某个指定session的事件处理,由某个指定worker负责
type poolRaceSelf struct {
	workers []*basic_event_queue.EventQueue

	closeDone chan struct{}
}

func (p *poolRaceSelf) TypeName() string {
	return poolName
}

func NewPoolRaceSelf() iface.WorkerPool {
	return &poolRaceSelf{
		workers:   make([]*basic_event_queue.EventQueue, workerNum),
		closeDone: make(chan struct{}),
	}
}

func (p *poolRaceSelf) Start() {
	for i := range p.workers {
		p.workers[i] = basic_event_queue.NewEventQueue(queueSize, true)
	}

	for _, w := range p.workers {
		w.Start()
	}
}

func (p *poolRaceSelf) Stop() (done <-chan struct{}) {
	chans := make([]<-chan struct{}, workerNum)

	for _, w := range p.workers {
		chans = append(chans, w.Stop())
	}

	go func() {
		for _, c := range chans {
			<-c
		}

		if p.closeDone != nil {
			close(p.closeDone)
		}
	}()

	return p.closeDone
}

func (p *poolRaceSelf) Put(session iface.NetSession, cb func()) {
	w := p.workers[session.ID()%workerNum]

	if err := w.Put(cb); err != nil {
		//TODO logs
		w.MustPut(cb)

	}
}

func (p *poolRaceSelf) TryPut(session iface.NetSession, cb func()) bool {

	w := p.workers[session.ID()%workerNum]

	if err := w.Put(cb); err != nil {
		return false
	}

	return true
}
