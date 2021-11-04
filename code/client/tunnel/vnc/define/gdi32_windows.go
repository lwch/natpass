package define

import "syscall"

var (
	libGdi32, _ = syscall.LoadLibrary("Gdi32.dll")
	// FuncGetDeviceCaps https://docs.microsoft.com/en-us/windows/win32/api/wingdi/nf-wingdi-getdevicecaps
	FuncGetDeviceCaps, _ = syscall.GetProcAddress(syscall.Handle(libGdi32), "GetDeviceCaps")
	// FuncCreateCompatibleDC https://docs.microsoft.com/en-us/windows/win32/api/wingdi/nf-wingdi-createcompatibledc
	FuncCreateCompatibleDC, _ = syscall.GetProcAddress(syscall.Handle(libGdi32), "CreateCompatibleDC")
	// FuncCreateCompatibleBitmap https://docs.microsoft.com/en-us/windows/win32/api/wingdi/nf-wingdi-createcompatiblebitmap
	FuncCreateCompatibleBitmap, _ = syscall.GetProcAddress(syscall.Handle(libGdi32), "CreateCompatibleBitmap")
	// FuncSelectObject https://docs.microsoft.com/en-us/windows/win32/api/wingdi/nf-wingdi-selectobject
	FuncSelectObject, _ = syscall.GetProcAddress(syscall.Handle(libGdi32), "SelectObject")
	// FuncDeleteDC https://docs.microsoft.com/en-us/windows/win32/api/wingdi/nf-wingdi-deletedc
	FuncDeleteDC, _ = syscall.GetProcAddress(syscall.Handle(libGdi32), "DeleteDC")
	// FuncDeleteObject https://docs.microsoft.com/en-us/windows/win32/api/wingdi/nf-wingdi-deleteobject
	FuncDeleteObject, _ = syscall.GetProcAddress(syscall.Handle(libGdi32), "DeleteObject")
	// FuncBitBlt https://docs.microsoft.com/en-us/windows/win32/api/wingdi/nf-wingdi-bitblt
	FuncBitBlt, _ = syscall.GetProcAddress(syscall.Handle(libGdi32), "BitBlt")
	// FuncGetDIBits https://docs.microsoft.com/en-us/windows/win32/api/wingdi/nf-wingdi-getdibits
	FuncGetDIBits, _ = syscall.GetProcAddress(syscall.Handle(libGdi32), "GetDIBits")
)

const (
	// HORZRES https://docs.microsoft.com/en-us/windows/win32/api/wingdi/nf-wingdi-getdevicecaps
	HORZRES = 8
	// VERTRES https://docs.microsoft.com/en-us/windows/win32/api/wingdi/nf-wingdi-getdevicecaps
	VERTRES = 10
	// BITSPIXEL https://docs.microsoft.com/en-us/windows/win32/api/wingdi/nf-wingdi-getdevicecaps
	BITSPIXEL = 12
)

const (
	// SRCCOPY https://docs.microsoft.com/en-us/windows/win32/api/wingdi/nf-wingdi-bitblt
	SRCCOPY = 0x00CC0020
)

const (
	// BI_RGB https://docs.microsoft.com/en-us/previous-versions/dd183376(v=vs.85)
	BI_RGB = 0
	// DIB_RGB_COLORS https://docs.microsoft.com/en-us/windows/win32/api/wingdi/nf-wingdi-getdibits
	DIB_RGB_COLORS = 0
)
