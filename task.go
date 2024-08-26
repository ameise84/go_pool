package go_pool

import (
	"sync"
)

type taskResult struct {
	result any
	err    error
}

type task struct {
	isSync   bool
	f        any
	args     []any
	syncChan chan taskResult
}

var taskPool sync.Pool

func init() {
	taskPool = sync.Pool{
		New: func() any {
			return &task{
				syncChan: make(chan taskResult),
			}
		},
	}
}

func newTask(isSync bool, f any, args ...any) *task {
	t := taskPool.Get().(*task)
	t.isSync = isSync
	t.f = f
	t.args = args
	return t
}

func freeTask(t *task) {
	if t != nil {
		t.f = nil
		t.args = nil
		taskPool.Put(t)
	}
}

func doTask(t *task) {
	if t.isSync {
		var r taskResult
		r.result, r.err = t.f.(GoSyncFunc)(t.args...)
		t.syncChan <- r
	} else {
		t.f.(GoAsyncFunc)(t.args...)
		freeTask(t) //异步的任务,所有权由执行者管理
	}
}

func doTaskSafe(h PanicHook, where string, t *task) {
	if t.isSync {
		var r taskResult
		r.err = ErrDoTaskPanic
		safeFunc(h, where, func() {
			r.result, r.err = t.f.(GoSyncFunc)(t.args...)
		})
		t.syncChan <- r
	} else {
		safeFunc(h, where, func() {
			t.f.(GoAsyncFunc)(t.args...)
		})
		freeTask(t) //异步的任务,所有权由执行者管理
	}
}
