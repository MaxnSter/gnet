package pool_norace

import (
	"sync"
	"time"

	"github.com/MaxnSter/gnet/pool"
)

const (
	Name                            = "poolNoRace"
	DefaultMaxGoroutinesAmount      = 256 * 1024
	DefaultMaxGoroutineIdleDuration = 10 * time.Second
)

func New() pool.Pool {
	return newPoolNoRace()
}

func init() {
	pool.RegisterWorkerPool(Name, New)
}

type goChan struct {
	lastUseTime time.Time
	ch          chan func()
}

// from fasthttp
type poolNoRace struct {
	maxGoroutinesAmount      int
	maxGoroutineIdleDuration time.Duration

	lock            *sync.Mutex
	goroutinesCount int
	mustStop        bool
	ready           []*goChan
	goChanPool      sync.Pool
	sync.Once

	stopCh    chan struct{}
	closeDone chan struct{}
}

// TypeName返回pool的唯一表示
func (p *poolNoRace) String() string {
	return Name
}

func newPoolNoRace() pool.Pool {
	return &poolNoRace{
		maxGoroutinesAmount:      DefaultMaxGoroutinesAmount,
		maxGoroutineIdleDuration: DefaultMaxGoroutineIdleDuration,
		lock:                     &sync.Mutex{},
		ready:                    make([]*goChan, 0),
		stopCh:                   make(chan struct{}),
		closeDone:                make(chan struct{}),
	}
}

func (p *poolNoRace) Run() {
	p.Once.Do(func() {
		p.start()
	})
}

func (p *poolNoRace) start() {
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
	ready, size, i := p.ready, len(p.ready), 0
	for i < size && curTime.Sub(ready[i].lastUseTime) >= p.maxGoroutineIdleDuration {
		i++
	}
	*scratch = append((*scratch)[:0], ready[:i]...)
	if i > 0 {
		m := copy(ready, ready[i:])
		for i := m; i < size; i++ {
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
			time.Sleep(1 * time.Second)

			if p.goroutinesCount == 0 {
				if p.closeDone != nil {
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
func (p *poolNoRace) Put(f func(), opts ...func(*pool.Option)) {
	select {
	case <-p.stopCh:
		return
	default:
	}

	if ch := p.getCh(); ch != nil {
		ch.ch <- f
	} else {
		f()
	}
}

// TryPut与Put相同,但当pool已满试,投放失败,返回false
func (p *poolNoRace) TryPut(f func(), opts ...func(*pool.Option)) bool {

	if ch := p.getCh(); ch != nil {
		ch.ch <- f
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
