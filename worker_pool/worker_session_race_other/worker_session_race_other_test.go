package worker_session_race_other

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/MaxnSter/gnet/iface"
	"github.com/stretchr/testify/assert"
)

type tSession struct {
	Id   int64
	race *int
}

func (ts *tSession) ID() int64                { return ts.Id }
func (ts *tSession) Send(message interface{}) {}
func (ts *tSession) Stop()                    {}

func (ts *tSession) Run() {
	for i := 0; i < math.MaxInt16; i++ {
	}
}

func TestNewPoolRaceOther(t *testing.T) {
	p := NewPoolRaceOther()
	assert.NotNil(t, p, "pool should not be nil")
	p.Start()

	ts := &tSession{
		Id:   1,
		race: new(int),
	}

	ts1 := &tSession{
		Id:   2,
		race: new(int),
	}

	wg := sync.WaitGroup{}
	fTs := func(_ iface.Context) {

		if ts.race != nil {
			*ts.race = 1
		}

		ts1.race = nil
		wg.Done()
	}

	fTs1 := func(_ iface.Context) {

		if ts1.race != nil {
			time.Sleep(time.Millisecond)
			*ts1.race = 1
		}
		ts.race = nil
		wg.Done()
	}

	for i := 0; i < 500000; i++ {

		wg.Add(1)

		//never panic
		if i%2 == 0 {
			p.Put(ts, fTs)
		} else {
			p.Put(ts, fTs1)
		}

		//may panic
		//if i%2 == 0 {
		//	go fTs()
		//} else {
		//	go fTs1()
		//}

	}

	wg.Wait()
}

func BenchmarkNewPoolRaceOther(b *testing.B) {

	p := NewPoolRaceOther()
	p.Start()

	ts := &tSession{
		Id:   1,
		race: new(int),
	}

	ts1 := &tSession{
		Id:   2,
		race: new(int),
	}

	wg := sync.WaitGroup{}
	fTs := func(_ iface.Context) {

		if ts.race != nil {
			*ts.race = 1
		}

		ts1.race = nil
		wg.Done()
	}

	fTs1 := func(_ iface.Context) {

		if ts1.race != nil {
			time.Sleep(time.Millisecond)
			*ts1.race = 1
		}
		ts.race = nil
		wg.Done()
	}

	for i := 0; i < 10000; i++ {

		wg.Add(1)

		//never panic
		if i%2 == 0 {
			p.Put(ts, fTs)
		} else {
			p.Put(ts, fTs1)
		}

		//MUST panic in Benchmark
		//if i%2 == 0 {
		//	go fTs()
		//} else {
		//	go fTs1()
		//}

	}

	wg.Wait()
}

func TestPoolRaceOther_Stop(t *testing.T) {

	q := NewPoolRaceOther()
	wg := &sync.WaitGroup{}
	q.Start()

	for i := 0; i < 5; i++ {
		wg.Add(1)
		q.Put(nil, func(_ iface.Context) {
			for i := 0; i < math.MaxUint8; i++ {
			}
			wg.Done()
		})
	}

	select {
	case <-time.After(10 * time.Second):
		assert.Fail(t, "queue1 not stopped")
	case <-q.StopAsync():
	}
}

func TestPoolRaceOther_Stop2(t *testing.T) {

	q := NewPoolRaceOther()
	wg := &sync.WaitGroup{}
	wgDoneCh := make(chan struct{})
	q.Start()

	for i := 0; i < 5; i++ {
		wg.Add(1)
		q.Put(nil, func(_ iface.Context) {
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
		assert.Fail(t, "queue1 stopped before all task finished!")
	case <-wgDoneCh:
	}
}
