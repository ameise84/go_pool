package go_pool

func safeFunc(h PanicHook, where string, f func()) {
	defer recoverPanic(h, where)
	f()
}
