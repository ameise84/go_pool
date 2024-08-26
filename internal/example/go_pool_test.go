package misc

import (
	"fmt"
	"github.com/ameise84/go_pool"

	"log"

	"runtime"
	"testing"
	"time"
)

const (
	testGoRunnerCount = 500000
	sleepTime         = 0 * time.Millisecond
)

func doFunc(args ...any) {
	x := args[0].(int)
	funcName := args[1].(string)
	if sleepTime > 0 {
		time.Sleep(sleepTime)
	}
	if x%10000 == 0 {
		log.Printf("[%s] do func x[%d]\n", funcName, x)
		panic("111")
	}
}

func doSyncFunc(args ...any) (any, error) {
	x := args[0].(int)
	funcName := args[1].(string)
	if sleepTime > 0 {
		time.Sleep(sleepTime)
	}
	if x%10000 == 0 {
		log.Printf("[%s] do func x[%d]\n", funcName, x)
		panic("111")
	}
	return nil, nil
}

type panicHandler struct {
}

func (p panicHandler) OnPanic(err error) {

}

func testCacheBlock() {
	runner := go_pool.NewGoRunner(panicHandler{}, "", go_pool.DefaultOptions().SetSimCount(8).SetBlock(true).SetCacheMode(true, 1024))
	for i := 0; i < testGoRunnerCount; i++ {
		x := i + 1
		if err := runner.AsyncRun(doFunc, x, "testCacheBlock"); err != nil {
			log.Printf("testCacheBlock err:%v\n", err)
		}
	}
	runner.Wait()
}

func testCacheNoBlock() {
	runner := go_pool.NewGoRunner(panicHandler{}, "", go_pool.DefaultOptions().SetSimCount(1).SetBlock(false).SetCacheMode(true, 16))
	succ, failed := 0, 0
	for i := 0; i < testGoRunnerCount; i++ {
		x := i + 1
		if err := runner.AsyncRun(doFunc, x, "testCacheNoBlock"); err != nil {
			failed++
			runtime.Gosched()
		} else {
			succ++
		}
	}
	runner.Wait()
	log.Printf("[testCacheNoBlock] do func:succ[%d] failed[%d]\n", succ, failed)
}

func testNoCacheBlock() {
	runner := go_pool.NewGoRunner(panicHandler{}, "", go_pool.DefaultOptions().SetSimCount(8).SetBlock(true).SetCacheMode(false, 0))
	for i := 0; i < testGoRunnerCount; i++ {
		x := i + 1
		if err := runner.AsyncRun(doFunc, x, "testNoCacheBlock"); err != nil {
			log.Printf("testNoCacheBlock err:%v\n", err)
		}
	}
	runner.Wait()
}

func testNoCacheNoBlock() {
	runner := go_pool.NewGoRunner(panicHandler{}, "", go_pool.DefaultOptions().SetSimCount(8).SetBlock(false).SetCacheMode(false, 0))
	succ, failed := 0, 0
	for i := 0; i < testGoRunnerCount; i++ {
		x := i + 1
		if err := runner.AsyncRun(doFunc, x, "testNoCacheNoBlock"); err != nil {
			failed++
			runtime.Gosched()
		} else {
			succ++
		}
	}
	runner.Wait()
	log.Printf("[testNoCacheNoBlock] do func:succ[%d] failed[%d]\n", succ, failed)
}

func testNoLimitSim() {
	runner := go_pool.NewGoRunner(panicHandler{}, "", go_pool.DefaultOptions().SetSimCount(0))
	for i := 0; i < testGoRunnerCount; i++ {
		x := i + 1
		if err := runner.AsyncRun(doFunc, x, "testNoLimitSim"); err != nil {
			log.Printf("testNoLimitSim err:%v\n", err)
		}
	}
	runner.Wait()
}

func testWaitFunc() {
	wf := go_pool.NewGoWaitFunc(panicHandler{})
	for i := 1; i <= testGoRunnerCount; i++ {
		wf.Run("", func() {
			doFunc(i, "testWaitFunc")
		})
	}
	wf.Wait()
}

func testGoWorker() {
	for i := 1; i <= testGoRunnerCount; i++ {
		go_pool.NewGoFuncDo(panicHandler{}, "", func() {
			doFunc(i, "testGoWorker")
		})
	}
}

func TestGoRunner(t *testing.T) {
	fmt.Println("---TestGoRunner---")
	testCacheBlock()
	log.Println("------------------------")
	testCacheNoBlock()
	log.Println("------------------------")
	testNoCacheBlock()
	log.Println("------------------------")
	testNoCacheNoBlock()
	log.Println("------------------------")
	testNoLimitSim()
	log.Println("------------------------")
	testWaitFunc()
	log.Println("------------------------")
	testGoWorker()
}
