package worker_session_race_other

import (
	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/gnet/worker_pool"
	"github.com/MaxnSter/gnet/worker_pool/internal/basic_event_queue"
)

const (
	poolName  = "poolRaceOther"
	queueSize = 1024
)

func init() {
	worker_pool.RegisterWorkerPool(poolName, NewPoolRaceOther)
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

func NewPoolRaceOther() worker_pool.Pool {
	return &poolRaceOther{
		queue: basic_event_queue.NewEventQueue(queueSize, true),
	}
}

func (p *poolRaceOther) Start() {
	logger.WithField("name", p.TypeName()).Infoln("worker_pool pool start")
	p.queue.Start()
}

func (p *poolRaceOther) StopAsync() (done <-chan struct{}) {
	logger.WithField("name", p.TypeName()).Infoln("pool stopping...")
	return p.queue.StopAsync()
}

func (p *poolRaceOther) Stop() {
	<-p.StopAsync()
	logger.WithField("name", p.TypeName()).Infoln("pool stopped...")
}

func (p *poolRaceOther) Put(ctx worker_pool.Context, cb func(worker_pool.Context)) {
	if err := p.queue.Put(ctx, cb); err != nil {
		//logger.WithField("name", p.TypeName()).Warningln("pool size limit")
		p.queue.MustPut(ctx, cb)
	}
}

func (p *poolRaceOther) TryPut(ctx worker_pool.Context, cb func(worker_pool.Context)) bool {

	if err := p.queue.Put(ctx, cb); err != nil {
		return false
	}

	return true
}
