package go_pool

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

func newNoCacheRunner(h PanicHook, where string, opts Options) *runnerNoCache {
	r := &runnerNoCache{
		opts:     opts,
		hand:     h,
		where:    where,
		taskChan: make(chan *task),
		simChan:  make(chan struct{}, opts.simCount),
	}
	return r
}

type runnerNoCache struct {
	doneTaskCnt atomic.Uint64
	opts        Options
	hand        PanicHook
	where       string
	taskChan    chan *task     //工作协程的chan
	simChan     chan struct{}  //控制并发数(控制协程数量)
	wg          sync.WaitGroup //等待所有任务处理完
}

func (r *runnerNoCache) Where() string {
	return r.where
}

func (r *runnerNoCache) post(t *task) (err error) {
	r.wg.Add(1)
	defer func() {
		if err != nil {
			r.wg.Done()
		}
	}()
	select {
	case r.taskChan <- t: //如果任务投递成功,则不开启新并发线程
	case r.simChan <- struct{}{}: //有可以使用的并发线程,则开启新线程执行任务
		r.newConsumer(t)
	default:
		if r.opts.simCount == 0 {
			r.newConsumer(t)
			return nil
		}
		if r.opts.isBlock {
			if r.opts.blockHand != nil {
				r.opts.blockHand.OnBlock(r)
			}
			if r.opts.blockOutTime > 0 {
				tr := time.NewTimer(r.opts.blockOutTime)
				select {
				case r.taskChan <- t:
				case <-tr.C:
					return ErrRunnerPostTimeOut
				}
				if !tr.Stop() {
					<-tr.C
				}
			} else {
				r.taskChan <- t
			}
		} else {
			return ErrRunnerWorkerIsBusy
		}
	}
	return nil
}

func (r *runnerNoCache) newConsumer(t *task) {
	w := _gPool.takeGoWorker()
	w.run(newTask(false, r.consume, t))
}

func (r *runnerNoCache) consume(args ...any) {
	doTaskSafe(r.hand, r.where, args[0].(*task))
	r.wg.Done()
	var tr *time.Timer
	if r.opts.goFreeTime == 0 {
		tr = &time.Timer{C: make(chan time.Time)}
	} else {
		tr = time.NewTimer(r.opts.goFreeTime)
	}
workLoop:
	for {
		select {
		case t := <-r.taskChan:
			if r.opts.goFreeTime > 0 {
				tr.Reset(r.opts.goFreeTime)
			}
			r.doneTaskCnt.Add(1)
			doTaskSafe(r.hand, r.where, t)
			r.wg.Done()
		case <-tr.C:
			break workLoop
		}
	}
	if r.opts.simCount > 0 {
		<-r.simChan
	}
}

func (r *runnerNoCache) SyncRun(f GoSyncFunc, args ...any) (any, error) {
	t := newTask(true, f, args...)
	defer freeTask(t)
	if err := r.post(t); err != nil {
		runtime.Gosched()
		return nil, err
	}
	rs := <-t.syncChan
	return rs.result, rs.err
}

func (r *runnerNoCache) AsyncRun(f GoAsyncFunc, args ...any) error {
	t := newTask(false, f, args...)
	if err := r.post(t); err != nil {
		freeTask(t)
		runtime.Gosched()
		return err
	}
	return nil
}

func (r *runnerNoCache) Wait() {
	r.wg.Wait()
}

func (r *runnerNoCache) DoneTaskCount() uint64 {
	return r.doneTaskCnt.Load()
}
