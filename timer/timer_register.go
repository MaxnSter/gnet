package timer

import "fmt"

type timerCreator func() Timer

var (
	timerCreators map[string]timerCreator
)

func RegisterTimer(name string, t timerCreator) {
	timerCreators[name] = t
}

func MustGetTimer(name string) Timer {
	if _, ok := timerCreators[name]; !ok {
		panic(fmt.Sprintf("timer:%s, not reister", name))
	}

	return timerCreators[name]()
}
