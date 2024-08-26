//go:build !debug

package go_pool

import (
	"errors"
	"fmt"
	"runtime/debug"
)

func recoverPanic(h PanicHook, where string) {
	if x := recover(); x != nil {
		stack := string(debug.Stack())
		msg := fmt.Sprintf("%+v[where:%s]\n stack:%s", x, where, stack)
		if h != nil {
			h.OnPanic(errors.New(msg))
		} else {
			_gLogger.Error(msg)
		}
	}
}
