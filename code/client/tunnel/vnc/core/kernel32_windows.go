package core

import "syscall"

var (
	libKernel32, _                      = syscall.LoadLibrary("kernel32.dll")
	funcWTSGetActiveConsoleSessionId, _ = syscall.GetProcAddress(syscall.Handle(libKernel32), "WTSGetActiveConsoleSessionId")
)

const PROCESS_ALL_ACCESS = 0x1F0FFF
