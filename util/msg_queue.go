package util

import (
	"context"
	"sync"
)

// MsgQueue 是一个双缓冲队列的BlockingQueue
// 在"多生产者,单消费者"的场景下下保证线程安全
type MsgQueue struct {
	list2       []interface{}
	list1       []interface{}
	produceList *[]interface{}

	lock   *sync.Mutex
	wakeup chan struct{}
}

// NewMsgQueue 创建并返回一个初始容量为0的消息队列
func NewMsgQueue() *MsgQueue {
	return NewMsgQueueWithCap(0)
}

// NewMsgQueueWithCap 创建并返回一个指定初始容量的消息队列
func NewMsgQueueWithCap(cap int) *MsgQueue {
	if cap < 0 {
		cap = 0
	}

	q := &MsgQueue{
		wakeup: make(chan struct{}, 1),
		list1:  make([]interface{}, 0, cap/2),
		list2:  make([]interface{}, 0, cap/2+cap%2),
		lock:   &sync.Mutex{},
	}

	q.lock.Lock()
	q.produceList = &q.list1
	q.lock.Unlock()
	return q
}

// Put 往队列中添加元素
func (q *MsgQueue) Put(msg interface{}) {
	q.lock.Lock()
	curList := q.produceList
	*curList = append(*curList, msg)
	q.lock.Unlock()

	//注意,此处有个竞态!
	//Add执行到此处时,len(*consumeList)正好为0但此处的select先执行,
	//pick就会一直阻塞.因此我们把wake channel size设为1
	//TODO 直接使用信号量
	select {
	case q.wakeup <- struct{}{}:
	default:
	}
}

// Pick 获取当前队列中的所有元素,若当前队列为空,则阻塞直至有新元素被添加
func (q *MsgQueue) Pick(retList *[]interface{}) {
	q.PickWithSignal(nil, retList)
}

// PickWithSignal 与Pick相同,但当signal active时,强制退出阻塞状态并返回
func (q *MsgQueue) PickWithSignal(signal <-chan struct{}, retList *[]interface{}) {
	q.lock.Lock()
	consumeList := q.consume()
	q.lock.Unlock()

	for len(*consumeList) == 0 {

		select {
		case <-signal:
			return
		case <-q.wakeup:

			q.lock.Lock()
			consumeList = q.consume()
			q.lock.Unlock()
		}
	}

	for i := 0; i < len(*consumeList); i++ {
		*retList = append(*retList, (*consumeList)[i])
		(*consumeList)[i] = nil
	}

	*consumeList = (*consumeList)[0:0]

}

// PickWithCtx 与Pick相同,但当ctx active时,强制退出阻塞状态并返回
func (q *MsgQueue) PickWithCtx(ctx context.Context, retList *[]interface{}) {
	q.PickWithSignal(ctx.Done(), retList)
}

// q.lock must locked
func (q *MsgQueue) consume() (consumeList *[]interface{}) {
	//一定程度上缩短了临界区长度
	if q.produceList == &q.list1 {
		consumeList = &q.list1
		q.produceList = &q.list2
	} else {
		consumeList = &q.list2
		q.produceList = &q.list1
	}

	return
}
