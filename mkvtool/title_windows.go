//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

func setWindowTitle(title string) {
	kernel32, err := syscall.LoadLibrary("kernel32.dll")
	if err == nil {
		defer syscall.FreeLibrary(kernel32)
		setConsoleTitle, err := syscall.GetProcAddress(kernel32, "SetConsoleTitleW")
		if err == nil {
			ptr, err := syscall.UTF16PtrFromString(title)
			if err == nil {
				syscall.SyscallN(setConsoleTitle, 1, uintptr(unsafe.Pointer(ptr)), 0, 0)
			}
		}
	}
}
