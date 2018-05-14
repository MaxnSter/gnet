package timer

import "time"

/*
This code is based on the following resources:
source code:	https://play.golang.org/p/Ys9qqanqmU
discuss:	https://groups.google.com/forum/#!msg/golang-dev/c9UUfASVPoU/tlbK2BpFEwAJ
*/
type safetimer struct {
	*time.Timer
	scr bool
}

//saw channel read, must be called after receiving value from safetimer chan
func (t *safetimer) SCR() {
	t.scr = true
}

func (t *safetimer) SafeReset(d time.Duration) bool {
	ret := t.Stop()

	//ret为true,表示timer并没有active,此时删除成功
	//ret为false,表示timer已经active,标准库time.Timer实现时,timer active时
	//其对应的callback的处理方式是(都是这个意思): go timer.callback()
	//所以此时,此时如果timer.C没有被处理(scr为true), 我们就需要手动drain timer.C
	if !ret && !t.scr {
		<-t.C
	}

	t.Timer.Reset(d)
	t.scr = false
	return ret
}

func newSafeTimer(d time.Duration) *safetimer {
	return &safetimer{
		Timer: time.NewTimer(d),
	}
}
