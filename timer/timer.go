package timer

import (
	"container/heap"
	"math"
	"sync"
	"time"

	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/gnet/util"
	"github.com/MaxnSter/gnet/worker_pool"
)

/*
一个基于最小堆的定时器
*/

//堆节点
type timerEntry struct {
	expire   time.Time     //到期时间
	interval time.Duration //重复间隔
	timerID  int64         //返回给用户的id
	index    int           //heap内部维护的index

	ctx iface.Context
	cb  OnTimeOut // callback
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
func (heap timerHeap) getTimerIdx(timerID int64) int {
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
)

var (
	_ TimerManager = (*timerManager)(nil)

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
	pool        worker_pool.Pool //负责处理callback的worker entryPool
	timers      timerHeap        //管理所有user timer的最小堆
	pauseCh     chan struct{}
	resumeCh    chan struct{}
	closeCh     chan struct{}
	closeDoneCh chan struct{}

	sync.Once
}

func newTimerManager(pool worker_pool.Pool) *timerManager {
	tm := &timerManager{
		pool:        pool,
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

// Start开启定时器功能,此方法保证goroutine safe
func (tm *timerManager) Start() {
	tm.Once.Do(func() {
		logger.Infoln("timer start")

		heap.Init(&tm.timers)
		go tm.run()
	})
}

// StopAsync关闭定时器,当定时器内部完全关闭时 done可读
func (tm *timerManager) StopAsync() (done <-chan struct{}) {
	tm.closeCh <- struct{}{}

	logger.Infoln("timer stopping...")
	return tm.closeDoneCh
}

// Stop关闭定时器,调用方阻塞直到定时器完全关闭
func (tm *timerManager) Stop() {
	<-tm.StopAsync()
}

// AddTimer添加一个定时任务,并返回该任务对应的id
func (tm *timerManager) AddTimer(expire time.Time, interval time.Duration, ctx iface.Context, cb OnTimeOut) (id int64) {
	t := tm.get()

	t.expire = expire
	t.interval = interval
	t.ctx = ctx
	t.cb = cb
	t.timerID = util.GetUUID()
	t.index = -1

	tm.pause()
	tm.timers.Push(t)
	tm.resume()

	return t.timerID
}

// CancelTimer取消一个定时
// node:如果该timer为一次性(interval = 0)且正好expire, 则取消无效
func (tm *timerManager) CancelTimer(id int64) {
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
		loopTimer    = newSafeTimer(UNTOUCHED)
		expiredTNode []*timerEntry
	)

	defer func() {
		loopTimer.Stop()
		close(tm.closeDoneCh)

		logger.Infoln("timer stopped")
	}()

	for {
		if len(tm.timers) > 0 {
			//选取一个最近时间的user timer
			timeout = tm.timers[0].expire.Sub(time.Now())
		} else {
			timeout = UNTOUCHED
		}
		loopTimer.safeReset(timeout)

		select {
		case <-tm.pauseCh:
			//wait for resume signal
			<-tm.resumeCh
		case <-tm.closeCh:
			return
		case <-loopTimer.Timer.C:
			loopTimer.scr()
			tm.expired(&expiredTNode)
		}
	}
}

//处理expired user timer的callback, 该func保证非阻塞执行
func (tm *timerManager) handleExpired(entry *timerEntry, t time.Time) {
	f := func(ctx iface.Context) {
		entry.cb(t, ctx)
	}

	if !tm.pool.TryPut(entry.ctx, f) {
		//must be async
		go tm.pool.Put(entry.ctx, f)
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
