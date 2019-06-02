package pool_race_other

import (
	"github.com/MaxnSter/gnet/pool"
	"github.com/MaxnSter/gnet/pool/plugins/internal/basic_event_queue"
)

const (
	Name      = "poolRaceOther"
	queueSize = 1024
)

func New() pool.Pool {
	return newPoolRaceOther()
}

func init() {
	pool.RegisterWorkerPool(Name, New)
}

//single EvnetLoop,保证绝对goroutine safe,可用于无锁服务
type poolRaceOther struct {
	queue *basic_event_queue.EventQueue
}

// TypeName返回pool的唯一表示
func (p *poolRaceOther) String() string {
	return Name
}

func newPoolRaceOther() pool.Pool {
	return &poolRaceOther{
		queue: basic_event_queue.NewEventQueue(queueSize),
	}
}

// Start启动pool,此方法保证goroutineeeee safe
func (p *poolRaceOther) Run() {
	p.queue.Run()
}

// StopAsync与Stop相同,但它立即返回, pool完全停止时done active
func (p *poolRaceOther) StopAsync() (done <-chan struct{}) {
	return p.queue.StopAsync()
}

// Stop停止pool,调用方阻塞直到Stop返回
// pool保证此时剩余的pool item全部执行完毕才返回
func (p *poolRaceOther) Stop() {
	<-p.StopAsync()
}

// Put往pool中投放任务,无论pool是否已满,此次投放必定成功
func (p *poolRaceOther) Put(f func(), opts ...func(*pool.Option)) {
	if ok := p.TryPut(f); !ok {
		p.queue.MustPut(f)
	}
}

// TryPut与Put相同,但当pool已满试,投放失败,返回false
func (p *poolRaceOther) TryPut(f func(), opts ...func(*pool.Option)) bool {

	if err := p.queue.Put(f); err != nil {
		return false
	}

	return true
}
