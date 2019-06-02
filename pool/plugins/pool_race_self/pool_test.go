package pool_race_self

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

func (ts *tSession) ID() int64                { return ts.Id }
func (ts *tSession) Send(message interface{}) {}
func (ts *tSession) Stop()                    {}

func TestNewPoolRaceSelf(t *testing.T) {
	p := newPoolRaceSelf()
	assert.NotNil(t, p, "pool should not be nil")
	p.Start()

	wg := sync.WaitGroup{}
	raceFunc := func(sId int) {
		nwg := sync.WaitGroup{}
		ts := &tSession{Id: int64(sId), race: new(int)}

		raceF := func(ctx iface.Context) {
			if ts.race != nil {
				time.Sleep(time.Millisecond)
				*ts.race = 1
			}
			nwg.Done()
		}

		raceF1 := func(ctx iface.Context) {
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

func TestPoolRaceSelf_Stop(t *testing.T) {

	q := newPoolRaceSelf()
	q.Start()

	for i := 0; i < 10; i++ {
		idx := i
		q.Put(&tSession{int64(idx), nil}, func(ctx iface.Context) {
			for i := 0; i < math.MaxUint8; i++ {
			}
			logger.WithField("i", idx).Infoln("task done")
		})
	}

	select {
	case <-time.After(5 * time.Second):
		assert.Fail(t, "queue not stopped")
	case <-q.StopAsync():
	}
}

func TestPoolRaceSelf_Stop2(t *testing.T) {

	q := newPoolRaceSelf()
	wg := &sync.WaitGroup{}
	wgDoneCh := make(chan struct{})
	q.Start()

	for i := 0; i < 5000; i++ {
		wg.Add(1)
		idx := i
		q.Put(&tSession{int64(idx), nil}, func(_ iface.Context) {
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
