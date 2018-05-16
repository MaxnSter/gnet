package timer

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/worker"
	_ "github.com/MaxnSter/gnet/worker/worker_session_norace"
	"github.com/stretchr/testify/assert"
)

var wPool iface.WorkerPool

func TestNewTimerManager(t *testing.T) {
	wPool = worker.MustGetWorkerPool("poolNoRace")
	wPool.Start()

	tw := NewTimerManager(wPool)
	assert.NotNil(t, tw, "tw should not be nil")
}

func TestTimerManager_AddTimer(t *testing.T) {
	wPool = worker.MustGetWorkerPool("poolNoRace")
	wPool.Start()

	tw := NewTimerManager(wPool)
	assert.NotNil(t, tw, "tw should not be nil")
	tw.Start()

	wg := sync.WaitGroup{}
	timerIds := make([]int64, 0)

	for i := 0; i < 10000; i++ {
		id := tw.AddTimer(time.Now(), time.Second, nil, func(i time.Time, session iface.NetSession) {
			wg.Add(1)
			for i := 0; i < math.MaxInt16; i++ {
			}
			wg.Done()
		})
		timerIds = append(timerIds, id)
	}

	var stopId int64
	wg.Add(1)
	stopId = tw.AddTimer(time.Now().Add(5*time.Second), 0, nil, func(i time.Time, session iface.NetSession) {
		tw.StopTimer(stopId)
		for _, id := range timerIds {
			tw.StopTimer(id)
		}
		wg.Done()
	})

	wg.Wait()
}

func TestTimerManager_Stop(t *testing.T) {
	wPool = worker.MustGetWorkerPool("poolNoRace")
	wPool.Start()

	tw := NewTimerManager(wPool)
	assert.NotNil(t, tw, "tw should not be nil")
	tw.Start()

	wg := sync.WaitGroup{}

	for i := 0; i < 10000; i++ {
		tw.AddTimer(time.Now(), time.Second, nil, func(i time.Time, session iface.NetSession) {
			wg.Add(1)
			for i := 0; i < math.MaxInt16; i++ {
			}
			wg.Done()
		})
	}

	var stopId int64
	wg.Add(1)
	stopId = tw.AddTimer(time.Now().Add(5*time.Second), 0, nil, func(i time.Time, session iface.NetSession) {
		tw.StopTimer(stopId)
		tw.Stop()
		wg.Done()
	})

	wg.Wait()
}
