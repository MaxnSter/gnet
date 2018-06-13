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
	worker_pool.RegisterWorkerPool(poolName, NewPoolNoRace)
}

//无data race情况

//from fastHttp
type goChan struct {
	lastUseTime time.Time
	ch          chan func()
}

type poolNoRace struct {
	maxGoroutinesAmount      int
	maxGoroutineIdleDuration time.Duration

	lock            *sync.Mutex
	goroutinesCount int
	mustStop        bool
	ready           []*goChan
	goChanPool      sync.Pool
	cbWrapper       basic_event_queue.CallBackWrapper

	stopCh    chan struct{}
	closeDone chan struct{}
}

func (p *poolNoRace) TypeName() string {
	return poolName
}

func NewPoolNoRace() worker_pool.Pool {
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

func (p *poolNoRace) Start() {
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

func (p *poolNoRace) StopAsync() (done <-chan struct{}) {

	select {
	case <-p.stopCh:
		panic("pool already stop")
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

func (p *poolNoRace) Stop() {
	<-p.StopAsync()
}

func (p *poolNoRace) Put(ctx worker_pool.Context, cb func(worker_pool.Context)) {

	select {
	case <-p.stopCh:
		logger.WithField("name", p.TypeName()).Errorln("pool already stopping, can't put task")
		return
	default:
	}

	if ch := p.getCh(); ch != nil {
		ch.ch <- basic_event_queue.Decorate(ctx, cb, p.cbWrapper)
	} else {
		logger.WithField("name", p.TypeName()).Warning("pool size limit")
		cb(ctx)
	}
}

func (p *poolNoRace) TryPut(ctx worker_pool.Context, cb func(worker_pool.Context)) bool {

	if ch := p.getCh(); ch != nil {
		ch.ch <- basic_event_queue.Decorate(ctx, cb, p.cbWrapper)
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
