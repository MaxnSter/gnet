package timer

import "time"

/*
This code is based on the following resources:
source code:	https://play.golang.org/p/Ys9qqanqmU
discuss:	https://groups.google.com/forum/#!msg/golang-dev/c9UUfASVPoU/tlbK2BpFEwAJ
*/
type safetimer struct {
	*time.Timer
	bScr bool
}

//saw channel read,在drain timer对应的ch后一定要调用
func (t *safetimer) scr() {
	t.bScr = true
}

//重置定时器
func (t *safetimer) safeReset(d time.Duration) bool {
	ret := t.Stop()

	//ret为true,表示timer并没有active,此时删除成功
	//ret为false,表示timer已经active,标准库time.Timer实现时,timer active时
	//其对应的callback的处理方式是(都是这个意思): go timer.callback()
	//所以此时,此时如果timer.C没有被处理(scr为true), 我们就需要手动drain timer.C
	if !ret && !t.bScr {
		<-t.C
	}

	t.Timer.Reset(d)
	t.bScr = false
	return ret
}

func newSafeTimer(d time.Duration) *safetimer {
	return &safetimer{
		Timer: time.NewTimer(d),
	}
}
