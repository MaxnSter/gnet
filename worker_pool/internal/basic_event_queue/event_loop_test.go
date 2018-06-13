package basic_event_queue

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/MaxnSter/gnet/iface"
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
		q.Put(nil, func(_ iface.Context) {
			for i := 0; i < math.MaxInt16; i++ {
			}
			wg.Done()
		})
	}

	wg.Wait()
}

func TestEventQueue_Put2(t *testing.T) {
	q := NewEventQueue(100, true)
	q.Start()
	defer q.Stop()

	q.Put(nil, func(_ iface.Context) {
		panic("panic")
	})

	time.Sleep(time.Second * 2)

}

func TestEventQueue_Stop(t *testing.T) {

	q := NewEventQueue(100, true)
	q.Start()

	for i := 0; i < 5000; i++ {
		q.MustPut(nil, func(_ iface.Context) {
			for i := 0; i < math.MaxUint8; i++ {
			}
		})
	}

	select {
	case <-time.After(10 * time.Second):
		assert.Fail(t, "queue not stopped")
	case <-q.StopAsync():

	}
}
