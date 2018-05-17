package worker_session_race_self

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/logger"
	"github.com/stretchr/testify/assert"
)

type tSession struct {
	Id   int64
	race *int
}

func (ts *tSession) ID() int64                  { return ts.Id }
func (ts *tSession) Send(message iface.Message) {}
func (ts *tSession) Stop()                      {}

func TestNewPoolRaceSelf(t *testing.T) {
	p := NewPoolRaceSelf()
	assert.NotNil(t, p, "pool should not be nil")
	p.Start()

	wg := sync.WaitGroup{}
	raceFunc := func(sId int) {
		nwg := sync.WaitGroup{}
		ts := &tSession{Id: int64(sId), race: new(int)}

		raceF := func() {
			if ts.race != nil {
				time.Sleep(time.Millisecond)
				*ts.race = 1
			}
			nwg.Done()
		}

		raceF1 := func() {
			ts.race = nil
			nwg.Done()
		}

		for i := 0; i < 500; i++ {
			nwg.Add(1)

			//may panic
			//if i%2 == 0 {
			//	go raceF()
			//} else {
			//	go raceF1()
			//}

			//never panic
			if i%2 == 0 {
				p.Put(ts, raceF)
			} else {
				p.Put(ts, raceF1)
			}

		}

		nwg.Wait()
		wg.Done()
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go raceFunc(i)
	}

	wg.Wait()
}

func BenchmarkNewPoolRaceSelf(b *testing.B) {

	p := NewPoolRaceSelf()
	p.Start()

	wg := sync.WaitGroup{}
	raceFunc := func(sId int) {
		nwg := sync.WaitGroup{}
		ts := &tSession{Id: int64(sId), race: new(int)}

		raceF := func() {
			if ts.race != nil {
				time.Sleep(time.Millisecond)
				*ts.race = 1
			}
			nwg.Done()
		}

		raceF1 := func() {
			ts.race = nil
			nwg.Done()
		}

		for i := 0; i < 500; i++ {
			nwg.Add(1)

			//MUST panic in benchmark
			//if i%2 == 0 {
			//	go raceF()
			//} else {
			//	go raceF1()
			//}

			//never panic
			if i%2 == 0 {
				p.Put(ts, raceF)
			} else {
				p.Put(ts, raceF1)
			}

		}

		nwg.Wait()
		wg.Done()
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go raceFunc(i)
	}

	wg.Wait()
}

func TestPoolRaceSelf_Stop(t *testing.T) {

	q := NewPoolRaceSelf()
	q.Start()

	for i := 0; i < 10; i++ {
		idx := i
		q.Put(&tSession{int64(idx), nil}, func() {
			for i := 0; i < math.MaxUint8; i++ {
			}
			logger.WithField("i", idx).Infoln("task done")
		})
	}

	select {
	case <-time.After(5 * time.Second):
		assert.Fail(t, "queue not stopped")
	case <-q.Stop():
	}
}

func TestPoolRaceSelf_Stop2(t *testing.T) {

	q := NewPoolRaceSelf()
	wg := &sync.WaitGroup{}
	wgDoneCh := make(chan struct{})
	q.Start()

	for i := 0; i < 500000; i++ {
		wg.Add(1)
		idx := i
		q.Put(&tSession{int64(idx), nil}, func() {
			for i := 0; i < math.MaxInt16; i++ {
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
	case <-time.After(60 * time.Second):
		assert.Fail(t, "queue stopped before all task finished!")
	case <-wgDoneCh:
	}
}
