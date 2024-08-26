package go_pool

import (
	"sync/atomic"
)

func goWorkerDo(f GoAsyncFunc, args ...any) {
	w := _gPool.takeGoWorker()
	w.run(newTask(false, f, args...))
}

func newGoWorker() any {
	w := &goWorker{
		signalChan: make(chan struct{}),
	}
	w.isRunning.Store(false)
	return w
}

type goWorker struct {
	t          *task
	signalChan chan struct{}
	isRunning  atomic.Bool
}

func (w *goWorker) run(t *task) {
	w.t = t
	if w.isRunning.CompareAndSwap(false, true) {
		go w.loop()
	}
	w.signalChan <- struct{}{}
}

func (w *goWorker) loop() {
	for {
		<-w.signalChan
		if w.t == nil {
			break //没有任务,直接关闭
		}
		doTask(w.t)
		w.t = nil
		_gPool.freeGoWorker(w)
	}
	w.isRunning.Store(false)
	_gPool.dropGoWorker(w)
}
