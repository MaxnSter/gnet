package util

import (
	"context"
	"sync"
)

// msgQueue是一个支持Add和Pick的消息队列
// 在"多生产者,单消费者"的模型下保证线程安全
type MsgQueue struct {
	list2       []interface{}
	list1       []interface{}
	produceList *[]interface{}

	lock   *sync.Mutex
	wakeup chan struct{}
}

// NewMsgQueue创建并返回一个容量为0的消息队列
func NewMsgQueue() *MsgQueue {
	return NewMsgQueueWithCap(0)
}

// NewMsgQueueWithCap创建并返回一个指定容量的消息队列
func NewMsgQueueWithCap(cap int) *MsgQueue {
	q := &MsgQueue{
		wakeup: make(chan struct{}, 0),
		list1:  make([]interface{}, 0, cap/2),
		list2:  make([]interface{}, 0, cap/2+cap%2),
		lock:   &sync.Mutex{},
	}

	q.lock.Lock()
	q.produceList = &q.list1
	q.lock.Unlock()
	//atomic.StorePointer(&q.produceList, unsafe.Pointer(&q.list1))
	return q
}

// Add往队列中添加元素
func (q *MsgQueue) Add(msg interface{}) {
	q.lock.Lock()
	curList := q.produceList
	//curList := (*[]interface{})(atomic.LoadPointer(&q.produceList))
	*curList = append(*curList, msg)
	q.lock.Unlock()

	select {
	case q.wakeup <- struct{}{}:
	default:
	}
}

// Pick获取当前队列中的所有元素,若当前队列为空,则阻塞直至有新元素被添加
func (q *MsgQueue) Pick(retList *[]interface{}) {

	q.lock.Lock()
	consumeList := q.consume()
	q.lock.Unlock()

	if len(*consumeList) == 0 {

		<-q.wakeup

		q.lock.Lock()
		consumeList = q.consume()
		q.lock.Unlock()
	}

	//在这里,consumeList与produceList一定不同,因此不需要上锁
	for i := 0; i < len(*consumeList); i++ {
		*retList = append(*retList, (*consumeList)[i])
		(*consumeList)[i] = nil
	}

	*consumeList = (*consumeList)[0:0]
}

// PickWithSignal与Pick相同,但当signal可读时,强制退出循环状态
func (q *MsgQueue) PickWithSignal(signal <-chan struct{}, retList *[]interface{}) {

	q.lock.Lock()
	consumeList := q.consume()
	q.lock.Unlock()

	if len(*consumeList) == 0 {

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

// PickWithSignal与Pick相同,但当ctx active时,强制退出循环状态
func (q *MsgQueue) PickWithCtx(ctx context.Context, retList *[]interface{}) {

	q.lock.Lock()
	consumeList := q.consume()
	q.lock.Unlock()

	if len(*consumeList) == 0 {

		select {
		case <-ctx.Done():
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

// q.lock must locked
func (q *MsgQueue) consume() (consumeList *[]interface{}) {

	if q.produceList == &q.list1 {
		consumeList = &q.list1
		q.produceList = &q.list2
	} else {
		consumeList = &q.list2
		q.produceList = &q.list1
	}

	return

	//for {
	//
	//	if atomic.CompareAndSwapPointer(
	//		&q.produceList,
	//		unsafe.Pointer(&q.list1),
	//		unsafe.Pointer(&q.list2)) {
	//
	//		consumeList = &q.list1
	//		return
	//	}
	//
	//	if atomic.CompareAndSwapPointer(
	//		&q.produceList,
	//		unsafe.Pointer(&q.list2),
	//		unsafe.Pointer(&q.list1)) {
	//
	//		consumeList = &q.list2
	//		return
	//	}
	//}

}
