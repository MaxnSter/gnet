package util

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMsgQueue(t *testing.T) {
	q := NewMsgQueueWithCap(100)
	assert.NotNil(t, q)
}

func TestMsgQueue_Add(t *testing.T) {

	var msgs []interface{}
	q := NewMsgQueueWithCap(100)
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {

		defer func() {
			wg.Done()
		}()

		for {

			q.Pick(&msgs)
			for _, msg := range msgs {

				if msg == nil {
					return
				}

				process(msg)
			}
			msgs = msgs[0:0]
		}
	}()

	go func() {
		for i := 0; i < 10000; i++ {
			q.Add(i)
		}
	}()

	go func() {
		for i := 0; i < 10000; i++ {
			q.Add(i)
		}
		q.Add(nil)
	}()

	go func() {
		for i := 0; i < 10000; i++ {
			q.Add(i)
		}
	}()
}

func process(msg interface{}) {
	time.Sleep(time.Millisecond)
}
