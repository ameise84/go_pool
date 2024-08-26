package go_pool

import (
	"sync"
)

type GoWaitFunc interface {
	Run(where string, cb func())
	Wait()
}

// GoWaitFunc 并发任务,等待所有任务完成
type goWaitFunc struct {
	hand PanicHook
	wg   sync.WaitGroup
}

func (w *goWaitFunc) Run(where string, cb func()) {
	w.wg.Add(1)
	wg := sync.WaitGroup{}
	wg.Add(1)
	goWorkerDo(func(...any) {
		wg.Done()
		safeFunc(w.hand, where, cb)
		w.wg.Done()
	})
	wg.Wait()
}

func (w *goWaitFunc) Wait() {
	w.wg.Wait()
}
