//go:build windows

package main

import (
    "syscall"
    "unsafe"
)

var (
    kernel32                       = syscall.NewLazyDLL("kernel32.dll")
    procGetStdHandle               = kernel32.NewProc("GetStdHandle")
    procGetConsoleMode             = kernel32.NewProc("GetConsoleMode")
    procSetConsoleMode             = kernel32.NewProc("SetConsoleMode")
)

const (
    stdOutputHandle                = ^uintptr(0) - 11 + 1 // STD_OUTPUT_HANDLE = -11
    enableVirtualTerminalProcessing = 0x0004
)

// enableWindowsColors enables ANSI color support in Windows terminal
func enableWindowsColors() {
    handle, _, _ := procGetStdHandle.Call(stdOutputHandle)
    if handle == 0 {
        return
    }

    var mode uint32
    procGetConsoleMode.Call(handle, uintptr(unsafe.Pointer(&mode)))
    mode |= enableVirtualTerminalProcessing
    procSetConsoleMode.Call(handle, uintptr(mode))
}
