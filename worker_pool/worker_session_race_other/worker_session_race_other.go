package worker_session_race_other

import (
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/gnet/worker_pool"
	"github.com/MaxnSter/gnet/worker_pool/internal/basic_event_queue"
)

const (
	poolName  = "poolRaceOther"
	queueSize = 1024
)

func init() {
	worker_pool.RegisterWorkerPool(poolName, newPoolRaceOther)
}

//single EvnetLoop,保证绝对goroutine safe,可用于无锁服务
type poolRaceOther struct {
	queue *basic_event_queue.EventQueue
}

// TypeName返回pool的唯一表示
func (p *poolRaceOther) TypeName() string {
	return poolName
}

func newPoolRaceOther() worker_pool.Pool {
	return &poolRaceOther{
		queue: basic_event_queue.NewEventQueue(queueSize, true),
	}
}

// Start启动pool,此方法保证goroutineeeee safe
func (p *poolRaceOther) Start() {
	logger.WithField("name", p.TypeName()).Infoln("worker_pool pool start")
	p.queue.Start()
}

// StopAsync与Stop相同,但它立即返回, pool完全停止时done active
func (p *poolRaceOther) StopAsync() (done <-chan struct{}) {
	logger.WithField("name", p.TypeName()).Infoln("pool stopping...")
	return p.queue.StopAsync()
}

// Stop停止pool,调用方阻塞直到Stop返回
// pool保证此时剩余的pool item全部执行完毕才返回
func (p *poolRaceOther) Stop() {
	<-p.StopAsync()
	logger.WithField("name", p.TypeName()).Infoln("pool stopped...")
}

// Put往pool中投放任务,无论pool是否已满,此次投放必定成功
func (p *poolRaceOther) Put(ctx iface.Context, cb func(iface.Context)) {
	if err := p.queue.Put(ctx, cb); err != nil {
		//logger.WithField("name", p.TypeName()).Warningln("pool size limit")
		p.queue.MustPut(ctx, cb)
	}
}

// TryPut与Put相同,但当pool已满试,投放失败,返回false
func (p *poolRaceOther) TryPut(ctx iface.Context, cb func(iface.Context)) bool {

	if err := p.queue.Put(ctx, cb); err != nil {
		return false
	}

	return true
}
