package util

import (
	"sync"
	"time"

	"github.com/MaxnSter/gnet/logger"
	"github.com/pkg/errors"
)

/*
	uuid生成器, from snowFlake
*/
const (
	workerBits uint8 = 10 //表示节点的位数,当前最多可以生成2^10=1024个节点
	numberBits uint8 = 12 //表示每个结点,1ms内可以生成的ID序列号,当前每1ms生成2^12=4096唯一ID

	workerMax   int64 = -1 ^ (-1 << workerBits)
	numberMax   int64 = -1 ^ (-1 << numberBits)
	timeShift   uint8 = workerBits + numberBits
	workerShift uint8 = numberBits

	// -|----------------41bit------------------|-----10bit-----|------12bit--------|
	//  |                                       |               |                   |
	//  |               时间戳i                  |	节点id       |      序列号       |
	//  |                                       |               |                   |
	// -|---------------------------------------|---------------|-------------------|

	//从上图可以看出,时间戳区域从0开始,可以使用 (2^41)/1000/3600/24/365 ~= 69年保证不重复
	//如果使用time.now() << 41 表示时间戳区域,那个时间戳区域就不是从0开始了,
	//而是从首次使用的时间开始,最多可以保证不重复至1970+69=2039年,
	//因此,使用 (time.now() - epoch) << 41
	epoch int64 = 1526624047
)

var gUUIDWorker *uuidWorker

func init() {
	gUUIDWorker, _ = NewUUIDWorker(0)
}

type uuidWorker struct {
	guard     *sync.Mutex
	timestamp int64
	workerId  int64
	number    int64 //当前毫秒已经生成的序列号的个数(从0开始累加)
}

//指定一个节点id,返回该节点对应的uuid生成器
//使用多个uuidWorker时,id由调用者保证不重复
func NewUUIDWorker(workerId int64) (*uuidWorker, error) {
	if workerId < 0 || workerId > workerMax {
		return nil, errors.New("error worker id")
	}

	return &uuidWorker{
		guard:    &sync.Mutex{},
		workerId: workerId,
	}, nil
}

//生成一个uuid
func (w *uuidWorker) GetUUID() int64 {
	w.guard.Lock()
	defer w.guard.Unlock()

try:
	curTime := time.Now().UnixNano() / 1e6
	if w.timestamp == curTime {
		w.number++

		//这一毫秒内生成的uuid数量已达到最大值,需要等待下一毫秒
		if w.number > numberMax {
			logger.Warningln("uuid's max number/ms limit, spin for 1ms")
			//自旋还是阻塞好?? benchmark测试后,spin比block性能好那么一点点
			for curTime <= w.timestamp {
				curTime = time.Now().UnixNano() / 1e6
			}
			goto try
		}
	} else {
		w.number = 0
		w.timestamp = curTime
	}

	// 见上文图
	return int64((w.timestamp-epoch)<<timeShift | (w.workerId << workerShift) | (w.number))
}

//gnet默认的uuidWorker生成一个uuid
func GetUUID() int64 {
	return gUUIDWorker.GetUUID()
}
