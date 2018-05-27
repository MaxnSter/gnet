package util

import (
	"sync/atomic"
	"unsafe"
)

type MsgQueue struct {
	list2       []interface{}
	list1       []interface{}
	produceList unsafe.Pointer

	wakeup chan struct{}
}

func NewMsgQueue() *MsgQueue {
	return NewMsgQueueWithCap(0)
}

func NewMsgQueueWithCap(cap int) *MsgQueue {
	q := &MsgQueue{
		wakeup: make(chan struct{}, 0),
		list1:  make([]interface{}, 0, cap/2),
		list2:  make([]interface{}, 0, cap/2+cap%2),
	}

	atomic.StorePointer(&q.produceList, unsafe.Pointer(&q.list1))
	return q
}

func (q *MsgQueue) Add(msg interface{}) {

	curList := (*[]interface{})(atomic.LoadPointer(&q.produceList))
	*curList = append(*curList, msg)

	select {
	case q.wakeup <- struct{}{}:
	default:
	}
}

func (q *MsgQueue) Pick(retList *[]interface{}, timeout <-chan struct{}) {

	consumeList := q.consume()

	if len(*consumeList) == 0 {

		select {
		case <-timeout:
			return
		case <-q.wakeup:
			consumeList = q.consume()
		}

	}

	for i := 0; i < len(*consumeList); i++ {
		*retList = append(*retList, (*consumeList)[i])
		(*consumeList)[i] = nil
	}

	*consumeList = (*consumeList)[0:0]
}

func (q *MsgQueue) consume() (consumeList *[]interface{}) {

	for {

		if atomic.CompareAndSwapPointer(
			&q.produceList,
			unsafe.Pointer(&q.list1),
			unsafe.Pointer(&q.list2)) {

			consumeList = &q.list1
			return
		}

		if atomic.CompareAndSwapPointer(
			&q.produceList,
			unsafe.Pointer(&q.list2),
			unsafe.Pointer(&q.list1)) {

			consumeList = &q.list2
			return
		}
	}

}
