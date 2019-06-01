package timer

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/pool"
	_ "github.com/MaxnSter/gnet/pool/pool_norace"
	"github.com/stretchr/testify/assert"
)

var wPool pool.Pool

func TestNewTimerManager(t *testing.T) {
	wPool = pool.MustGetWorkerPool("poolNoRace")
	wPool.Start()

	tw := NewWithPool(wPool)
	assert.NotNil(t, tw, "tw should not be nil")
}

func TestTimerManager_AddTimer(t *testing.T) {
	wPool = pool.MustGetWorkerPool("poolNoRace")
	wPool.Start()

	tw := NewWithPool(wPool)
	assert.NotNil(t, tw, "tw should not be nil")
	tw.Start()

	wg := sync.WaitGroup{}
	timerIds := make([]int64, 0)

	for i := 0; i < 10000; i++ {
		id := tw.AddTimer(time.Now(), time.Second, nil, func(i time.Time, ctx iface.Context) {
			wg.Add(1)
			for i := 0; i < math.MaxInt16; i++ {
			}
			wg.Done()
		})
		timerIds = append(timerIds, id)
	}

	var stopId int64
	wg.Add(1)
	stopId = tw.AddTimer(time.Now().Add(5*time.Second), 0, nil, func(i time.Time, ctx iface.Context) {
		tw.CancelTimer(stopId)
		for _, id := range timerIds {
			tw.CancelTimer(id)
		}
		wg.Done()
	})

	wg.Wait()
}

func TestTimerManager_Stop(t *testing.T) {
	wPool = pool.MustGetWorkerPool("poolNoRace")
	wPool.Start()

	tw := NewWithPool(wPool)
	assert.NotNil(t, tw, "tw should not be nil")
	tw.Start()

	wg := sync.WaitGroup{}

	for i := 0; i < 10000; i++ {
		tw.AddTimer(time.Now(), time.Second, nil, func(i time.Time, ctx iface.Context) {
			wg.Add(1)
			for i := 0; i < math.MaxInt16; i++ {
			}
			wg.Done()
		})
	}

	var stopId int64
	wg.Add(1)
	stopId = tw.AddTimer(time.Now().Add(5*time.Second), 0, nil, func(i time.Time, ctx iface.Context) {
		tw.CancelTimer(stopId)
		tw.Stop()
		wg.Done()
	})

	wg.Wait()
}
