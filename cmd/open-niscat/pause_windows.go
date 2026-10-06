package main

import (
	"bufio"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var getConsoleProcessList = syscall.NewLazyDLL("kernel32.dll").NewProc("GetConsoleProcessList")

// pauseIfOwnConsole keeps the window open after an error when the program was started by double-click
// (it is then the only process attached to its console), so the message can be read.
func pauseIfOwnConsole() {
	var pids [2]uint32
	n, _, _ := getConsoleProcessList.Call(uintptr(unsafe.Pointer(&pids[0])), uintptr(len(pids))) //nolint:gosec // Win32 call with a fixed-size local buffer
	if n != 1 {
		return
	}
	fmt.Fprint(os.Stderr, "Press Enter to close this window.")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}
