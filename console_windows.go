//go:build windows

package main

import "syscall"

// init sets the Windows console output to UTF-8 (code page 65001) so emojis and
// special characters render correctly instead of showing as "?" or boxes.
func init() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setConsoleOutputCP := kernel32.NewProc("SetConsoleOutputCP")
	const cpUTF8 = 65001
	setConsoleOutputCP.Call(uintptr(cpUTF8))
}
