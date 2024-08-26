//go:build debug

package go_pool

// recoverPanic debug 模式不抓取panic
func recoverPanic(PanicHook, string) {
}
