package pool_race_self

import (
	"math/rand"
	"time"

	"github.com/MaxnSter/gnet/pool"
	"github.com/MaxnSter/gnet/pool/plugins/internal/basic_event_queue"
)

const (
	poolName = "poolRaceSelf"

	workerNum = 20
	queueSize = 256
)

func New() pool.Pool {
	return newPoolRaceSelf()
}

func init() {
	rand.Seed(time.Now().UnixNano())
	pool.RegisterWorkerPool(poolName, newPoolRaceSelf)
}

//session存在data race现象,并且几乎没有其他session交互的情况
//此时某个指定session的事件处理,由某个指定worker负责,worker的选择方式为round robin
type poolRaceSelf struct {
	workers []*basic_event_queue.EventQueue

	closeDone chan struct{}
}

func (p *poolRaceSelf) String() string {
	return poolName
}

func newPoolRaceSelf() pool.Pool {
	return &poolRaceSelf{
		workers:   make([]*basic_event_queue.EventQueue, workerNum),
		closeDone: make(chan struct{}),
	}
}

// Start启动pool,此方法保证goroutineeeee safe
func (p *poolRaceSelf) Run() {
	for i := range p.workers {
		p.workers[i] = basic_event_queue.NewEventQueue(queueSize)
	}

	for _, w := range p.workers {
		w.Run()
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

	for _, w := range p.workers {
		chans = append(chans, w.StopAsync())
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

// Put往pool中投放任务,无论pool是否已满,此次投放必定成功
func (p *poolRaceSelf) Put(f func(), opts ...func(*pool.Option)) {
	var w *basic_event_queue.EventQueue
	o := &pool.Option{}
	for _, f := range opts {
		f(o)
	}

	if o.Identifier != nil {
		w = p.workers[o.Identifier.ID()%workerNum]
	} else {
		w = p.workers[rand.Intn(workerNum)%workerNum]
	}

	if err := w.Put(f); err != nil {
		w.MustPut(f)
	}
}

// TryPut与Put相同,但当pool已满试,投放失败,返回false
func (p *poolRaceSelf) TryPut(f func(), opts ...func(*pool.Option)) bool {
	var w *basic_event_queue.EventQueue
	o := &pool.Option{}
	for _, f := range opts {
		f(o)
	}

	if o.Identifier != nil {
		w = p.workers[o.Identifier.ID()%workerNum]
	} else {
		w = p.workers[rand.Intn(workerNum)%workerNum]
	}

	if err := w.Put(f); err != nil {
		return false
	}
	return true
}
