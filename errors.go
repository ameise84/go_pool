package go_pool

import "errors"

var (
	ErrDoTaskPanic        = errors.New("go worker do syncChan task panic")
	ErrRunnerPostTimeOut  = errors.New("go runner post time out")
	ErrRunnerWorkerIsBusy = errors.New("go runner worker is busy")
)
