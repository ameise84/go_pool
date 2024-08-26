package go_pool

import "sync"

var _gPool goWorkerPool

func init() {
	_gPool = goWorkerPool{
		workerCache: make(chan *goWorker, 128), //协程缓存数量
		pool:        sync.Pool{New: newGoWorker},
	}
}

func setGoWorkerCacheCount(cnt uint16) {
	newCache := make(chan *goWorker, cnt)
loopFor:
	for {
		select {
		case w := <-_gPool.workerCache:
			newCache <- w
		default:
			break loopFor
		}
	}
	close(_gPool.workerCache)
	_gPool.workerCache = newCache
	return
}

type goWorkerPool struct {
	workerCache chan *goWorker
	pool        sync.Pool
}

func (p *goWorkerPool) takeGoWorker() *goWorker {
	select {
	case w := <-p.workerCache:
		return w
	default:
		w := _gPool.pool.Get().(*goWorker)
		return w
	}
}

func (p *goWorkerPool) freeGoWorker(w *goWorker) {
	select {
	case p.workerCache <- w:
	default:
		w.run(nil) // 利用无任务关闭协程
	}
}

func (p *goWorkerPool) dropGoWorker(w *goWorker) {
	_gPool.pool.Put(w)
}
