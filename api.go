package go_pool

type GoSyncFunc = func(...any) (any, error)

type GoAsyncFunc = func(...any)

type BlockHandler interface {
	OnBlock(GoRunner)
}

type PanicHook interface {
	OnPanic(err error)
}

type GoRunner interface {
	Where() string
	// SyncRun 执行任务,直到该任务被执行返回
	SyncRun(GoSyncFunc, ...any) (any, error)
	// AsyncRun 异步执行任务
	AsyncRun(GoAsyncFunc, ...any) error
	// Wait 等待所有任务处理完成
	Wait()
	//DoneTaskCount 已完成任务数量
	DoneTaskCount() uint64
}

// SetWorkerCacheCount 设置go_pool缓存大小
func SetWorkerCacheCount(cnt uint16) {
	setGoWorkerCacheCount(cnt)
}

// NewGoRunner runner是对Worker的包装,用于执行复杂且长时间需要处理的任务. runner可以设定并发能力及任务队列长度
func NewGoRunner(h PanicHook, where string, opts ...*Options) GoRunner {
	var o = DefaultOptions()
	if opts != nil && opts[0] != nil {
		o = opts[0]
	}
	if o.isCache {
		return newCacheRunner(h, where, *o)
	} else {
		return newNoCacheRunner(h, where, *o)
	}
}

// NewGoFuncDo 通常及其简单的任务,且无需等待成功执行完成的,通过此方法使用
func NewGoFuncDo(h PanicHook, where string, f func()) {
	goWorkerDo(func(...any) {
		safeFunc(h, where, f)
	})
}

// NewGoWaitFunc 和NewGoWorkerDo类似，但可以等待任务完成,这个和无并发限制的GoRunner类似，只是所有任务都是异步执行的，同时必须等待所有task执行完成
func NewGoWaitFunc(h PanicHook) GoWaitFunc {
	return &goWaitFunc{hand: h}
}
