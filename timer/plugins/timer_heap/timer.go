package timer_heap

import (
	"container/heap"
	"github.com/MaxnSter/gnet/timer"
	"github.com/MaxnSter/gnet/timer/plugins/internal"
	"math"
	"sync"
	"time"

	"github.com/MaxnSter/gnet/pool"
	"github.com/MaxnSter/gnet/util"
)

/*
一个基于最小堆的定时器
*/

//堆节点
type timerEntry struct {
	expire   time.Time     //到期时间
	interval time.Duration //重复间隔
	timerID  uint64
	index    int //heap内部维护的index

	cb timer.OnTimeOut // callback
}

type timerHeap []*timerEntry

//实现标准库heap接口
func (heap timerHeap) Len() int           { return len(heap) }
func (heap timerHeap) Less(i, j int) bool { return heap[i].expire.Sub(heap[j].expire) < 0 }
func (heap timerHeap) Swap(i, j int) {
	heap[i], heap[j] = heap[j], heap[i]
	heap[i].index, heap[j].index = i, j
}

func (heap *timerHeap) Push(x interface{}) {
	n := len(*heap)
	tn := x.(*timerEntry)
	tn.index = n
	*heap = append(*heap, tn)
}

func (heap *timerHeap) Pop() interface{} {
	old := *heap
	n := len(old)
	tn := old[n-1]
	tn.index = -1
	*heap = old[0 : n-1]
	return tn
}

//根据timerId来获取该timerId对应的timer在堆内的index
func (heap timerHeap) getTimerIdx(timerID uint64) int {
	for i := range heap {
		if heap[i].timerID == timerID {
			return heap[i].index
		}
	}

	return -1
}

const (
	//dummy duration,堆中不存在任何user timer时
	//此参数作为sys timer的参数,这样可以一直阻塞
	UNTOUCHED = time.Duration(math.MaxInt64)

	Name = "timer_heap"
)

var (
	_ timer.Timer = (*timerManager)(nil)

	ePool = &entryPool{
		p: sync.Pool{
			New: func() interface{} {
				return new(timerEntry)
			}},
	}
)

type entryPool struct {
	p sync.Pool
}

func (ep *entryPool) get() *timerEntry {
	return ep.p.Get().(*timerEntry)
}

func (ep *entryPool) put(t *timerEntry) {
	ep.p.Put(t)
}

type timerManager struct {
	pool        pool.Pool //负责处理callback的worker entryPool
	timers      timerHeap //管理所有user timer的最小堆
	pauseCh     chan struct{}
	resumeCh    chan struct{}
	closeCh     chan struct{}
	closeDoneCh chan struct{}

	sync.Once
}

func init() {
	timer.RegisterTimer(Name, New)
}

func New() timer.Timer {
	return newTimerManager()
}

func newTimerManager() timer.Timer {
	tm := &timerManager{
		timers:      make([]*timerEntry, 0),
		pauseCh:     make(chan struct{}),
		resumeCh:    make(chan struct{}),
		closeCh:     make(chan struct{}),
		closeDoneCh: make(chan struct{}),
	}

	return tm
}

func (tm *timerManager) put(t *timerEntry) {
	ePool.put(t)
}

func (tm *timerManager) get() *timerEntry {
	return ePool.get()
}

func (tm *timerManager) String() string {
	return Name
}

func (tm *timerManager) SetPool(p pool.Pool) {
	tm.pool = p
}

func (tm *timerManager) Run() {
	tm.Once.Do(func() {

		heap.Init(&tm.timers)
		go tm.run()
	})
}

func (tm *timerManager) stopAsync() (done <-chan struct{}) {
	tm.closeCh <- struct{}{}

	return tm.closeDoneCh
}

// Stop关闭定时器,调用方阻塞直到定时器完全关闭
func (tm *timerManager) Stop() {
	<-tm.stopAsync()
}

// AddTimer添加一个定时任务,并返回该任务对应的id
func (tm *timerManager) AddTimer(expire time.Time, interval time.Duration, cb timer.OnTimeOut) timer.Cancel {
	t := tm.get()

	t.expire = expire
	t.interval = interval
	t.cb = cb
	t.timerID = util.GetUUID()
	t.index = -1

	tm.pause()
	tm.timers.Push(t)
	tm.resume()

	return func() {
		tm.CancelTimer(t.timerID)
	}
}

// CancelTimer取消一个定时
// node:如果该timer为一次性(interval = 0)且正好expire, 则取消无效
func (tm *timerManager) CancelTimer(id uint64) {
	idx := tm.timers.getTimerIdx(id)
	if idx == -1 {
		return
	}

	tm.pause()
	t := heap.Remove(&tm.timers, idx)
	tm.resume()

	tm.put(t.(*timerEntry))
}

func (tm *timerManager) pause() {
	tm.pauseCh <- struct{}{}
}

func (tm *timerManager) resume() {
	tm.resumeCh <- struct{}{}
}

func (tm *timerManager) run() {
	var (
		timeout      time.Duration
		loopTimer    = internal.NewSafeTimer(UNTOUCHED)
		expiredTNode []*timerEntry
	)

	defer func() {
		loopTimer.Stop()
		close(tm.closeDoneCh)
	}()

	for {
		if len(tm.timers) > 0 {
			//选取一个最近时间的user timer
			timeout = tm.timers[0].expire.Sub(time.Now())
		} else {
			timeout = UNTOUCHED
		}
		loopTimer.SafeReset(timeout)

		select {
		case <-tm.pauseCh:
			//wait for resume signal
			<-tm.resumeCh
		case <-tm.closeCh:
			return
		case <-loopTimer.Timer.C:
			loopTimer.Scr()
			tm.expired(&expiredTNode)
		}
	}
}

//处理expired user timer的callback, 该func保证非阻塞执行
func (tm *timerManager) handleExpired(entry *timerEntry, t time.Time) {
	f := func() {
		entry.cb(t)
	}

	if !tm.pool.TryPut(f) {
		//must be async
		go tm.pool.Put(f)
	}
}

//有一个或多个user timer 到期
func (tm *timerManager) expired(expiredTNode *[]*timerEntry) {
	for len(tm.timers) > 0 {
		t := heap.Pop(&tm.timers).(*timerEntry)
		if time.Since(t.expire) > 0 {
			*expiredTNode = append(*expiredTNode, t)
			tm.handleExpired(t, time.Now())
		} else {
			heap.Push(&tm.timers, t)
			break
		}
	}

	tm.update(*expiredTNode)
	for i := range *expiredTNode {
		(*expiredTNode)[i] = nil
	}
	*expiredTNode = (*expiredTNode)[:0]
}

//对所有interval不为0的user timer,调整expire time,重新入堆
func (tm *timerManager) update(tNodes []*timerEntry) {

	if len(tNodes) == 0 {
		return
	}

	for i, v := range tNodes {
		if v.interval <= 0 {
			tm.put(tNodes[i])
			continue
		}

		v.expire = time.Now().Add(v.interval)
		heap.Push(&tm.timers, tNodes[i])
	}
}
