package timer

import (
	"container/heap"
	"math"
	"sync/atomic"
	"time"

	"github.com/MaxnSter/gnet/iface"
)

type timerNode struct {
	expire   time.Time
	interval time.Duration
	timerId  int64
	index    int //heap内部维护

	session iface.NetSession
	cb      TimeOutCB
}

type timerHeap []*timerNode

//heap
func (heap timerHeap) Len() int           { return len(heap) }
func (heap timerHeap) Less(i, j int) bool { return heap[i].expire.Sub(heap[j].expire) < 0 }
func (heap timerHeap) Swap(i, j int) {
	heap[i], heap[j] = heap[j], heap[i]
	heap[i].index, heap[j].index = i, j
}

func (heap *timerHeap) Push(x interface{}) {
	n := len(*heap)
	tn := x.(*timerNode)
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

func (heap timerHeap) getTimerIdx(timerId int64) int {
	for i := range heap {
		if heap[i].timerId == timerId {
			return heap[i].index
		}
	}

	return -1
}

type TimeOutCB func(time.Time, iface.NetSession)

const (
	_UNTOUCHED = time.Duration(math.MaxInt64)
)

type timerManager struct {
	workers    iface.WorkerPool
	timers     timerHeap
	timerIdGen int64
	pauseCh    chan struct{}
	resumeCh   chan struct{}
	closeCh    chan struct{}
}

func NewTimerManager(workers iface.WorkerPool) *timerManager {
	return &timerManager{
		workers:  workers,
		timers:   make([]*timerNode, 0),
		pauseCh:  make(chan struct{}),
		resumeCh: make(chan struct{}),
		closeCh:  make(chan struct{}),
	}
}

func (tm *timerManager) Start() {
	heap.Init(&tm.timers)
	go tm.run()
}

func (tm *timerManager) Stop() {
	//TODO return a <-chan ??
	tm.closeCh <- struct{}{}
}

func (tm *timerManager) AddTimer(expire time.Time, interval time.Duration, cb TimeOutCB) (id int64) {
	t := &timerNode{
		expire:   expire,
		interval: interval,
		cb:       cb,
		timerId:  atomic.AddInt64(&tm.timerIdGen, 1),
	}

	tm.pause()
	tm.timers.Push(tm)
	tm.resume()

	return t.timerId
}

func (tm *timerManager) StopTimer(id int64) {
	idx := tm.timers.getTimerIdx(id)
	if idx == -1 {
		return
	}

	tm.pause()
	heap.Remove(&tm.timers, idx)
	tm.resume()
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
		loopTimer    = newSafeTimer(_UNTOUCHED)
		expiredTNode []*timerNode
	)

	defer loopTimer.Stop()

	for {
		if len(tm.timers) > 0 {
			timeout = tm.timers[0].expire.Sub(time.Now())
		} else {
			timeout = _UNTOUCHED
		}
		loopTimer.Reset(timeout)

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

func (tm *timerManager) handleExpired(tNode *timerNode, t time.Time) {
	f := func() {
		tNode.cb(t, tNode.session)
	}

	if !tm.workers.TryPut(tNode.session, f) {
		//must be async
		//TODO warning
		go tm.workers.Put(tNode.session, f)
	}
}

func (tm *timerManager) expired(expiredTNode *[]*timerNode) {
	for len(tm.timers) > 0 {
		t := heap.Pop(&tm.timers).(*timerNode)
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

func (tm *timerManager) update(tNodes []*timerNode) {

	if len(tNodes) == 0 {
		return
	}

	for i, v := range tNodes {
		if v.interval <= 0 {
			continue
		}

		//TODO which?
		//v.expire = v.expire.Add(v.interval)
		v.expire = time.Now().Add(v.interval)
		heap.Push(&tm.timers, tNodes[i])
	}
}
