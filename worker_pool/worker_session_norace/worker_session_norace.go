package worker_session_norace

import (
	"sync"
	"time"

	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/gnet/worker_pool"
	"github.com/MaxnSter/gnet/worker_pool/internal/basic_event_queue"
	"github.com/sirupsen/logrus"
)

const (
	poolName                        = "poolNoRace"
	DefaultMaxGoroutinesAmount      = 256 * 1024
	DefaultMaxGoroutineIdleDuration = 10 * time.Second
)

func init() {
	worker_pool.RegisterWorkerPool(poolName, newPoolNoRace)
}

type goChan struct {
	lastUseTime time.Time
	ch          chan func()
}

// 此pool可用于无data race情况,常用于无状态服务
// 摘自fasthttp的pool,一个可伸缩的教科书级别的goroutine池
type poolNoRace struct {
	maxGoroutinesAmount      int
	maxGoroutineIdleDuration time.Duration

	lock            *sync.Mutex
	goroutinesCount int
	mustStop        bool
	ready           []*goChan
	goChanPool      sync.Pool
	cbWrapper       basic_event_queue.CallBackWrapper
	sync.Once

	stopCh    chan struct{}
	closeDone chan struct{}
}

// TypeName返回pool的唯一表示
func (p *poolNoRace) TypeName() string {
	return poolName
}

func newPoolNoRace() worker_pool.Pool {
	return &poolNoRace{
		maxGoroutinesAmount:      DefaultMaxGoroutinesAmount,
		maxGoroutineIdleDuration: DefaultMaxGoroutineIdleDuration,
		lock:      &sync.Mutex{},
		ready:     make([]*goChan, 0),
		stopCh:    make(chan struct{}),
		closeDone: make(chan struct{}),
		cbWrapper: basic_event_queue.SafeCallBack,
	}
}

// Start启动pool,此方法保证goroutineeeee safe
func (p *poolNoRace) Start() {
	p.Once.Do(func() {
		p.start()
	})
}

func (p *poolNoRace) start() {
	logger.WithField("name", p.TypeName()).Infoln("worker_pool pool start")
	go func() {
		var scratch []*goChan
		for {
			select {
			case <-p.stopCh:
				return
			default:
				time.Sleep(p.maxGoroutineIdleDuration)
				p.clean(&scratch)
			}
		}
	}()
}

func (p *poolNoRace) clean(scratch *[]*goChan) {
	curTime := time.Now()

	p.lock.Lock()
	ready, len, i := p.ready, len(p.ready), 0
	for i < len && curTime.Sub(ready[i].lastUseTime) >= p.maxGoroutineIdleDuration {
		i++
	}
	*scratch = append((*scratch)[:0], ready[:i]...)
	if i > 0 {
		m := copy(ready, ready[i:])
		for i := m; i < len; i++ {
			ready[i] = nil
		}
		p.ready = p.ready[:m]
	}
	p.lock.Unlock()

	tmp := *scratch
	for i, ch := range tmp {
		ch.ch <- nil
		tmp[i] = nil
	}
}

// StopAsync与Stop相同,但它立即返回, pool完全停止时done active
func (p *poolNoRace) StopAsync() (done <-chan struct{}) {

	select {
	case <-p.stopCh:
		return
	default:
	}

	logger.WithField("name", p.TypeName()).Infoln("pool stopping...")
	close(p.stopCh)

	p.lock.Lock()
	for i := range p.ready {
		p.ready[i].ch <- nil
		p.ready[i] = nil
	}
	p.ready = p.ready[:0]
	p.mustStop = true
	p.lock.Unlock()

	go func() {
		for {
			time.Sleep(3 * time.Second)

			logger.WithFields(logrus.Fields{"name": p.TypeName(), "size": p.goroutinesCount}).
				Infoln("waiting for workers exit...")

			if p.goroutinesCount == 0 {
				if p.closeDone != nil {
					logger.WithField("name", p.TypeName()).Infoln("pool stopped")
					close(p.closeDone)
				}
				return
			}
		}
	}()

	return p.closeDone
}

// Stop停止pool,调用方阻塞直到Stop返回
// pool保证此时剩余的pool item全部执行完毕才返回
func (p *poolNoRace) Stop() {
	<-p.StopAsync()
}

// Put往pool中投放任务,无论pool是否已满,此次投放必定成功
func (p *poolNoRace) Put(ctx iface.Context, cb func(iface.Context)) {

	select {
	case <-p.stopCh:
		logger.WithField("name", p.TypeName()).Errorln("pool already stopping, can't put task")
		return
	default:
	}

	if ch := p.getCh(); ch != nil {
		ch.ch <- basic_event_queue.Bind(ctx, cb, p.cbWrapper)
	} else {
		logger.WithField("name", p.TypeName()).Warning("pool size limit")
		cb(ctx)
	}
}

// TryPut与Put相同,但当pool已满试,投放失败,返回false
func (p *poolNoRace) TryPut(ctx iface.Context, cb func(iface.Context)) bool {

	if ch := p.getCh(); ch != nil {
		ch.ch <- basic_event_queue.Bind(ctx, cb, p.cbWrapper)
		return true
	}

	return false
}

func (p *poolNoRace) getCh() *goChan {
	var ch *goChan
	createGoroutine := false

	p.lock.Lock()
	ready, n := p.ready, len(p.ready)-1
	if n < 0 {
		if p.goroutinesCount < p.maxGoroutinesAmount {
			createGoroutine = true
			p.goroutinesCount++
		}
	} else {
		ch = ready[n]
		ready[n] = nil
		p.ready = p.ready[:n]
	}
	p.lock.Unlock()

	//get a goChan from freeList
	if ch != nil {
		return ch
	}

	//maxGoroutineAmount limit
	if !createGoroutine {
		return nil
	}

	vch := p.goChanPool.Get()
	if vch == nil {
		vch = &goChan{
			ch:          make(chan func(), 1),
			lastUseTime: time.Now(),
		}
	}
	ch = vch.(*goChan)

	go func() {
		p.goroutineFunc(ch)

		p.goChanPool.Put(vch)
	}()

	return ch
}

func (p *poolNoRace) goroutineFunc(goCh *goChan) {
	for cb := range goCh.ch {
		if cb == nil {
			break
		}
		cb()
		if !p.release(goCh) {
			break
		}

	}

	p.lock.Lock()
	p.goroutinesCount--
	p.lock.Unlock()
}

func (p *poolNoRace) release(goCh *goChan) bool {
	goCh.lastUseTime = time.Now()

	p.lock.Lock()
	if p.mustStop {
		p.lock.Unlock()
		return false
	}
	p.ready = append(p.ready, goCh)
	p.lock.Unlock()

	return true
}
