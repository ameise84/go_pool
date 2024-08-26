package go_pool

import (
	"fmt"
	"os"
	"time"
)

type Logger interface {
	Error(any)
}

func SetLogger(log Logger) {
	_gLogger = log
}

const (
	errorFormat = "\033[91;1m[%s ERRO] %v\033[0m\n"
	layout      = "2006-01-02 15:04:05.000"
)

var (
	_gLogger Logger
)

func init() {
	_gLogger = logger{}
}

type logger struct {
}

func (l logger) Error(msg any) {
	_, _ = fmt.Fprintf(os.Stdout, errorFormat, time.Now().Format(layout), msg)
}
