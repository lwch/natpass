package define

import "syscall"

var (
	libKernel32, _                      = syscall.LoadLibrary("kernel32.dll")
	FuncWTSGetActiveConsoleSessionId, _ = syscall.GetProcAddress(syscall.Handle(libKernel32), "WTSGetActiveConsoleSessionId")
	FuncGlobalAlloc, _                  = syscall.GetProcAddress(syscall.Handle(libKernel32), "GlobalAlloc")
	FuncGlobalFree, _                   = syscall.GetProcAddress(syscall.Handle(libKernel32), "GlobalFree")
)

const PROCESS_ALL_ACCESS = 0x1F0FFF

const (
	GHND          = 0x0042
	GMEM_FIXED    = 0x0000
	GMEM_MOVEABLE = 0x0002
	GMEM_ZEROINIT = 0x0040
	GPTR          = 0x0040
)
