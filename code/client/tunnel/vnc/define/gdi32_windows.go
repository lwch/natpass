package define

import "syscall"

var (
	libGdi32, _                   = syscall.LoadLibrary("Gdi32.dll")
	FuncGetDeviceCaps, _          = syscall.GetProcAddress(syscall.Handle(libGdi32), "GetDeviceCaps")
	FuncCreateCompatibleDC, _     = syscall.GetProcAddress(syscall.Handle(libGdi32), "CreateCompatibleDC")
	FuncCreateCompatibleBitmap, _ = syscall.GetProcAddress(syscall.Handle(libGdi32), "CreateCompatibleBitmap")
	FuncSelectObject, _           = syscall.GetProcAddress(syscall.Handle(libGdi32), "SelectObject")
	FuncDeleteDC, _               = syscall.GetProcAddress(syscall.Handle(libGdi32), "DeleteDC")
	FuncDeleteObject, _           = syscall.GetProcAddress(syscall.Handle(libGdi32), "DeleteObject")
	FuncBitBlt, _                 = syscall.GetProcAddress(syscall.Handle(libGdi32), "BitBlt")
	FuncGetDIBits, _              = syscall.GetProcAddress(syscall.Handle(libGdi32), "GetDIBits")
)

const (
	HORZRES   = 8
	VERTRES   = 10
	BITSPIXEL = 12
)

const (
	SRCCOPY = 0x00CC0020
)

const (
	BI_RGB         = 0
	DIB_RGB_COLORS = 0
)
