package worker_session_norace

import (
	"sync"
	"time"

	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/worker"
)

const (
	poolName                        = "poolNoRace"
	DefaultMaxGoroutinesAmount      = 256 * 1024
	DefaultMaxGoroutineIdleDuration = 10 * time.Second
)

func init() {
	worker.RegisterWorkerPool(poolName, NewPoolNoRace)
}

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

	stopCh    chan struct{}
	closeDone chan struct{}
}

func (p *poolNoRace) TypeName() string {
	return poolName
}

func NewPoolNoRace() iface.WorkerPool {
	return &poolNoRace{
		maxGoroutinesAmount:      DefaultMaxGoroutinesAmount,
		maxGoroutineIdleDuration: DefaultMaxGoroutineIdleDuration,
		lock:   &sync.Mutex{},
		ready:  make([]*goChan, 0),
		stopCh: make(chan struct{}),
	}
}

func (p *poolNoRace) Start() {
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

func (p *poolNoRace) Stop() (done <-chan struct{}) {

	if p.stopCh == nil {
		panic("pool already stop")
	}
	close(p.stopCh)
	p.stopCh = nil

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
			select {
			case <-time.After(1 * time.Second):
				p.lock.Lock()
				if p.goroutinesCount == 0 {
					p.lock.Unlock()
					if p.closeDone != nil {
						close(p.closeDone)
					}
					return
				}
				p.lock.Unlock()
			}
		}
	}()

	return p.closeDone
}

func (p *poolNoRace) Put(session iface.NetSession, cb func()) {
	if ch := p.getCh(); ch != nil {
		ch.ch <- cb
	} else {
		//TODO warning
		cb()
	}
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

		if !p.release(p) {
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
