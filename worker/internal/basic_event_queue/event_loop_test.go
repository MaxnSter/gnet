package basic_event_queue

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewEventQueue(t *testing.T) {
	q := NewEventQueue(100, true)
	assert.NotNil(t, q, "queue should not be nil!")
}

func TestEventQueue_Put(t *testing.T) {
	q := NewEventQueue(100, true)
	wg := &sync.WaitGroup{}
	q.Start()

	for i := 0; i < 100; i++ {
		wg.Add(1)
		q.Put(func() {
			for i := 0; i <  math.MaxInt16; i++{}
			wg.Done()
		})
	}

	wg.Wait()
}

func TestEventQueue_Put2(t *testing.T) {
	q := NewEventQueue(100, true)
	q.Start()
	defer q.Stop()

	q.Put(func() {
		panic("panic")
	})

	time.Sleep(time.Second * 2)

}

func TestEventQueue_Stop(t *testing.T) {

	q := NewEventQueue(100, true)
	wg := &sync.WaitGroup{}
	wgDoneCh := make(chan struct{})
	q.Start()

	for i := 0; i < 100; i++ {
		wg.Add(1)
		q.Put(func() {
			for i := 0; i <  math.MaxUint8; i++{}
			wg.Done()
		})
	}

	q.Stop()
	wg.Add(1)
	q.Put(func() {
		wg.Done()
	})

	go func() {
		wg.Wait()
		wgDoneCh <- struct{}{}
	}()

	select {
	case <- time.After(10 * time.Second):
	case <- wgDoneCh:
		assert.Fail(t, "queue not stopped")
	}
}