package go_pool

import (
	"time"
)

func DefaultOptions() *Options {
	return &Options{
		isBlock:      true,
		blockHand:    nil,
		blockOutTime: 0,
		simCount:     1,
		isCache:      false,
		cacheSize:    0,
		goFreeTime:   10 * time.Second,
	}
}

type Options struct {
	isBlock      bool          // 提交任务时,是否阻塞到提交成功
	blockHand    BlockHandler  // 当阻塞时警告回调
	blockOutTime time.Duration // 阻塞超时设置
	simCount     uint32        // 并发数量,0:代表无限并发
	isCache      bool          // 是否开启消息缓存
	cacheSize    int32         // 消息缓存数量,0:代表无限缓存
	goFreeTime   time.Duration // 释放协程间隔
}

// SetBlock 设置阻塞模式
func (o *Options) SetBlock(isBlock bool, hand ...BlockHandler) *Options {
	o.isBlock = isBlock
	if hand != nil && hand[0] != nil {
		o.blockHand = hand[0]
	}
	return o
}

// SetBlockHandler 设置阻塞时的回调函数
func (o *Options) SetBlockHandler(f BlockHandler) *Options {
	o.blockHand = f
	return o
}

// SetBlockOutTime 设置投递任务最大阻塞时间
func (o *Options) SetBlockOutTime(t time.Duration) *Options {
	if !o.isBlock {
		panic("the block state is false, please set block true first")
		return o
	}
	if t < 0 {
		panic("the block out time is smaller than zero")
		return o
	}
	o.blockOutTime = t
	return o
}

// SetSimCount 设置最大并发数量
func (o *Options) SetSimCount(simCount uint32) *Options {
	if simCount == 0 && o.isCache {
		panic("the cache mode has been turned on. please turn off it")
		return o
	}
	o.simCount = simCount
	return o
}

// SetCacheMode 设置消息缓存
func (o *Options) SetCacheMode(isCache bool, cacheSize int32) *Options {
	if isCache {
		if o.simCount == 0 {
			panic("the goroutine mode is unlimited, please set simCount to non-zero")
		}
		if cacheSize < 0 {
			panic("the cache size is smaller than zero")
		}
	}

	o.isCache = isCache
	o.cacheSize = cacheSize
	return o
}

// SetGoFreeTime 设置goroutine释放时间
func (o *Options) SetGoFreeTime(t time.Duration) *Options {
	if t < 10*time.Second && t != 0 {
		t = 10 * time.Second
	}
	o.goFreeTime = t
	return o
}
