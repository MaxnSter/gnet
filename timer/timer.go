package timer

import (
	"container/heap"
	"math"
	"sync"
	"time"

	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/gnet/util"
)

/*
一个基于最小堆的定时器
*/

//堆节点
type timerEntry struct {
	expire   time.Time     //到期时间
	interval time.Duration //重复间隔
	timerId  int64         //返回给用户的id
	index    int           //heap内部维护的index

	//TODO 去掉该字段
	session iface.NetSession //该user timer对应的调用者
	cb      iface.TimeOutCB  // callback

	next *timerEntry
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
func (heap timerHeap) getTimerIdx(timerId int64) int {
	for i := range heap {
		if heap[i].timerId == timerId {
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
	_ iface.Timer = (*timerManager)(nil)
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
	workers     iface.WorkerPool //负责处理callback的worker pool
	timers      timerHeap        //管理所有user timer的最小堆
	pauseCh     chan struct{}
	resumeCh    chan struct{}
	closeCh     chan struct{}
	closeDoneCh chan struct{}

	pool sync.Pool
}

//指定一个worker pool,用于负责调用caller传入的callback,返回timerManager
//因为我们只用一个go routine来负责所有用户定时器,所以在timer expire时,
//对callback的处理,必须是非阻塞的.因此,把callback的调用,交给worker pool来做
//timerManager只负责往pool里填东西
func NewTimerManager(workers iface.WorkerPool) *timerManager {
	tm := &timerManager{
		workers:     workers,
		timers:      make([]*timerEntry, 0),
		pauseCh:     make(chan struct{}),
		resumeCh:    make(chan struct{}),
		closeCh:     make(chan struct{}),
		closeDoneCh: make(chan struct{}),
	}

	tm.pool.New = func() interface{} {
		return new(timerEntry)
	}

	return tm
}

func (tm *timerManager) put(t *timerEntry) {
	tm.pool.Put(t)
}

func (tm *timerManager) get() *timerEntry {
	return tm.pool.Get().(*timerEntry)
}

//开启定时器功能
func (tm *timerManager) Start() {
	logger.Infoln("timer start")

	heap.Init(&tm.timers)
	go tm.run()
}

func (tm *timerManager) StopAsync() (done <-chan struct{}) {
	tm.closeCh <- struct{}{}

	logger.Infoln("timer stopping...")
	return tm.closeDoneCh
}

//关闭定时器,当定时器内部完全关闭时 done可读
func (tm *timerManager) Stop() {
	tm.closeCh <- struct{}{}

	logger.Infoln("timer stopping...")
	<-tm.closeDoneCh
}

//添加一个定时
//TODO 接口优化,目前接口太丑陋了
func (tm *timerManager) AddTimer(expire time.Time, interval time.Duration, s iface.NetSession, cb iface.TimeOutCB) (id int64) {
	t := tm.get()

	t.expire = expire
	t.interval = interval
	t.session = s
	t.cb = cb
	t.timerId = util.GetUUID()
	t.index = -1

	tm.pause()
	tm.timers.Push(t)
	tm.resume()

	return t.timerId
}

//取消一个定时,如果该timer为一次性(interval = 0)正好expire,
//则取消无效
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
		loopTimer.SafeReset(timeout)

		select {
		case <-tm.pauseCh:
			//wait for resume signal
			<-tm.resumeCh
		case <-tm.closeCh:
			return
		case <-loopTimer.Timer.C:
			loopTimer.SCR()
			tm.expired(&expiredTNode)
		}
	}
}

//处理expired user timer的callback, 该func保证非阻塞执行
func (tm *timerManager) handleExpired(tNode *timerEntry, t time.Time) {
	f := func() {
		tNode.cb(t, tNode.session)
	}

	if !tm.workers.TryPut(tNode.session, f) {
		//must be async
		go tm.workers.Put(tNode.session, f)
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
