package define

import "syscall"

var (
	libGdi32, _          = syscall.LoadLibrary("Gdi32.dll")
	FuncGetDeviceCaps, _ = syscall.GetProcAddress(syscall.Handle(libGdi32), "GetDeviceCaps")
)

const (
	HORZRES   = 8
	VERTRES   = 10
	BITSPIXEL = 12
)
