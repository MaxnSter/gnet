package worker_session_norace

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type tSession struct{ Id int64 }

func (ts *tSession) ID() int64                { return ts.Id }
func (ts *tSession) Send(message interface{}) {}
func (ts *tSession) Stop()                    {}

func (ts *tSession) Run() {
	for i := 0; i < math.MaxInt16; i++ {
	}
}

func TestNewPoolNoRace(t *testing.T) {
	p := NewPoolNoRace()
	assert.NotNil(t, p, "pool should not be nil")

	ts := &tSession{Id: 1}
	wg := sync.WaitGroup{}
	p.Start()

	for i := 0; i < 500000; i++ {
		wg.Add(1)
		p.Put(ts, func() {
			ts.Run()
			wg.Done()
		})
	}

	p.Stop()
	wg.Wait()
}

func TestPoolNoRace_Stop(t *testing.T) {

	q := NewPoolNoRace()
	wg := &sync.WaitGroup{}
	wgDoneCh := make(chan struct{})
	q.Start()

	for i := 0; i < 10000; i++ {
		wg.Add(1)
		q.Put(nil, func() {
			for i := 0; i < math.MaxUint8; i++ {
			}
			wg.Done()
		})
	}

	q.Stop()
	wg.Add(1)
	q.Put(nil, func() {
		wg.Done()
	})

	go func() {
		wg.Wait()
		wgDoneCh <- struct{}{}
	}()

	select {
	case <-time.After(10 * time.Second):
	case <-wgDoneCh:
		assert.Fail(t, "queue not stopped")
	}
}

func TestPoolNoRace_Stop2(t *testing.T) {

	q := NewPoolNoRace()
	wg := &sync.WaitGroup{}
	wgDoneCh := make(chan struct{})
	q.Start()

	for i := 0; i < 5; i++ {
		wg.Add(1)
		q.Put(nil, func() {
			for i := 0; i < math.MaxInt8; i++ {
			}
			wg.Done()
		})
	}

	q.Stop()

	go func() {
		wg.Wait()
		wgDoneCh <- struct{}{}
	}()

	select {
	case <-time.After(6 * time.Second):
		assert.Fail(t, "queue stopped before all task finished!")
	case <-wgDoneCh:
	}
}
