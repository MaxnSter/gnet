package worker_session_race_self

import (
	"math/rand"

	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/gnet/worker_pool"
	"github.com/MaxnSter/gnet/worker_pool/internal/basic_event_queue"
)

const (
	poolName = "poolRaceSelf"

	workerNum = 20
	queueSize = 256
)

func init() {
	worker_pool.RegisterWorkerPool(poolName, newPoolRaceSelf)
}

//session存在data race现象,并且几乎没有其他session交互的情况
//此时某个指定session的事件处理,由某个指定worker负责,worker的选择方式为round robin
type poolRaceSelf struct {
	workers []*basic_event_queue.EventQueue

	closeDone chan struct{}
}

// TypeName返回pool的唯一表示
func (p *poolRaceSelf) TypeName() string {
	return poolName
}

func newPoolRaceSelf() worker_pool.Pool {
	return &poolRaceSelf{
		workers:   make([]*basic_event_queue.EventQueue, workerNum),
		closeDone: make(chan struct{}),
	}
}

// Start启动pool,此方法保证goroutineeeee safe
func (p *poolRaceSelf) Start() {
	logger.WithField("name", p.TypeName()).Infoln("worker_pool pool start")
	for i := range p.workers {
		p.workers[i] = basic_event_queue.NewEventQueue(queueSize, true)
	}

	for _, w := range p.workers {
		w.Start()
	}
}

// Stop停止pool,调用方阻塞直到Stop返回
// pool保证此时剩余的pool item全部执行完毕才返回
func (p *poolRaceSelf) Stop() {
	<-p.StopAsync()
}

// StopAsync与Stop相同,但它立即返回, pool完全停止时done active
func (p *poolRaceSelf) StopAsync() (done <-chan struct{}) {
	chans := make([]<-chan struct{}, 0, workerNum)

	logger.WithField("name", p.TypeName()).Infoln("pool stopping...")

	for _, w := range p.workers {
		chans = append(chans, w.StopAsync())
	}

	go func() {
		for _, c := range chans {
			<-c
		}

		if p.closeDone != nil {
			logger.WithField("name", p.TypeName()).Infoln("pool stopped")
			close(p.closeDone)
		}
	}()

	return p.closeDone
}

// Put往pool中投放任务,无论pool是否已满,此次投放必定成功
func (p *poolRaceSelf) Put(ctx iface.Context, cb func(iface.Context)) {
	var w *basic_event_queue.EventQueue

	if identifier, ok := ctx.(iface.Identifier); !ok {
		//TODO warning?
		w = p.workers[rand.Intn(workerNum)%workerNum]
	} else {
		w = p.workers[identifier.ID()%workerNum]
	}

	if err := w.Put(ctx, cb); err != nil {
		logger.WithField("name", p.TypeName()).Warning("pool size limit")
		w.MustPut(ctx, cb)
	}
}

// TryPut与Put相同,但当pool已满试,投放失败,返回false
func (p *poolRaceSelf) TryPut(ctx iface.Context, cb func(iface.Context)) bool {
	var w *basic_event_queue.EventQueue

	if identifier, ok := ctx.(iface.Identifier); !ok {
		//TODO warning?
		w = p.workers[rand.Intn(workerNum)%workerNum]
	} else {
		w = p.workers[identifier.ID()%workerNum]
	}

	if err := w.Put(ctx, cb); err != nil {
		return false
	}
	return true
}
