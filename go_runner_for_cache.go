package go_pool

import (
	"github.com/ameise84/lock"
	"github.com/ameise84/queue"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

func newCacheRunner(h PanicHook, where string, opts Options) *runnerCache {
	r := &runnerCache{
		opts:      opts,
		hand:      h,
		where:     where,
		taskChan:  make(chan struct{}, opts.simCount),
		simChan:   make(chan struct{}, opts.simCount),
		blockChan: make(chan struct{}, 1),
	}
	if opts.cacheSize == 0 {
		r.taskQueue = queue.NewListQueueLock[*task](&lock.SpinLock{})
	} else {
		r.taskQueue = queue.NewRingQueueLock[*task](opts.cacheSize, &lock.SpinLock{})
	}
	return r
}

type runnerCache struct {
	doneTaskCnt atomic.Uint64
	opts        Options
	hand        PanicHook
	where       string
	taskChan    chan struct{}
	simChan     chan struct{}
	blockChan   chan struct{} //不能使用sync.Cond.在投递失败时,在进入等待前,消费者优先投递通知消息会造成生产者被阻塞
	taskQueue   queue.IQueue[*task]
	wg          sync.WaitGroup //等待所有任务处理完
}

func (r *runnerCache) Where() string {
	return r.where
}

func (r *runnerCache) post(t *task) (err error) {
	r.wg.Add(1)
	defer func() {
		if err != nil {
			r.wg.Done()
		}
	}()

postAgain:
	if r.taskQueue.Enqueue(t) == nil {
		select {
		case r.simChan <- struct{}{}: //有可以使用的并发线程,则开启新线程执行任务
			r.newConsumer()
		default:
			select {
			case r.taskChan <- struct{}{}: //发送任务信号
			default:
			}
		}
	} else {
		if r.opts.isBlock {
			if r.opts.blockHand != nil {
				r.opts.blockHand.OnBlock(r)
			}
			if r.opts.blockOutTime > 0 {
				tr := time.NewTimer(r.opts.blockOutTime)
				select {
				case <-r.blockChan:
				case <-tr.C:
					return ErrRunnerPostTimeOut
				}
				if !tr.Stop() {
					<-tr.C
				}
			} else {
				<-r.blockChan
			}
			goto postAgain
		} else {
			return ErrRunnerWorkerIsBusy
		}
	}
	return nil
}

func (r *runnerCache) newConsumer() {
	goWorkerDo(r.consume)
}

func (r *runnerCache) consume(...any) {
	r.loopDoTask() //必须先执行一次,不能在newConsumer的时候通过signalChan通知
	var tr *time.Timer
	if r.opts.goFreeTime == 0 {
		tr = &time.Timer{C: make(chan time.Time)}
	} else {
		tr = time.NewTimer(r.opts.goFreeTime)
	}
workLoop:
	for {
		select {
		case <-r.taskChan:
			if r.opts.goFreeTime > 0 {
				tr.Reset(r.opts.goFreeTime)
			}
			r.loopDoTask()
		case <-tr.C:
			break workLoop
		}
	}
	<-r.simChan
}

func (r *runnerCache) loopDoTask() {
	for {
		t, err := r.taskQueue.Dequeue()
		select {
		case r.blockChan <- struct{}{}: //尝试唤醒生产者
		default:
		}
		if err != nil {
			break
		}

		r.doneTaskCnt.Add(1)
		doTaskSafe(r.hand, r.where, t)
		r.wg.Done()
	}
}

func (r *runnerCache) SyncRun(f GoSyncFunc, args ...any) (any, error) {
	t := newTask(true, f, args...)
	defer freeTask(t) //同步任务,所有权自行管理
	if err := r.post(t); err != nil {
		runtime.Gosched()
		return nil, err
	}
	rs := <-t.syncChan
	return rs.result, rs.err
}

func (r *runnerCache) AsyncRun(f GoAsyncFunc, args ...any) error {
	t := newTask(false, f, args...) //异步的任务,所有权由执行者管理
	if err := r.post(t); err != nil {
		freeTask(t) //投递失败,所有权还在自己
		runtime.Gosched()
		return err
	}
	return nil
}

func (r *runnerCache) Wait() {
	r.wg.Wait()
}

func (r *runnerCache) DoneTaskCount() uint64 {
	return r.doneTaskCnt.Load()
}
