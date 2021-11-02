package define

import "syscall"

var (
	libUser32, _            = syscall.LoadLibrary("user32.dll")
	FuncOpenInputDesktop, _ = syscall.GetProcAddress(syscall.Handle(libUser32), "OpenInputDesktop")
	FuncSetThreadDesktop, _ = syscall.GetProcAddress(syscall.Handle(libUser32), "SetThreadDesktop")
	FuncGetThreadDesktop, _ = syscall.GetProcAddress(syscall.Handle(libUser32), "GetThreadDesktop")
	FuncCloseDesktop, _     = syscall.GetProcAddress(syscall.Handle(libUser32), "CloseDesktop")
	FuncGetDesktopWindow, _ = syscall.GetProcAddress(syscall.Handle(libUser32), "GetDesktopWindow")
	FuncGetDC, _            = syscall.GetProcAddress(syscall.Handle(libUser32), "GetDC")
	FuncReleaseDC, _        = syscall.GetProcAddress(syscall.Handle(libUser32), "ReleaseDC")
	FuncGetCursorInfo, _    = syscall.GetProcAddress(syscall.Handle(libUser32), "GetCursorInfo")
	FuncGetIconInfo, _      = syscall.GetProcAddress(syscall.Handle(libUser32), "GetIconInfo")
	FuncDrawIcon, _         = syscall.GetProcAddress(syscall.Handle(libUser32), "DrawIcon")
)

type (
	HANDLE  uintptr
	BOOL    int32
	DWORD   uint32
	LONG    int32
	HCURSOR HANDLE
	HBITMAP HANDLE
)

type POINT struct {
	X LONG
	Y LONG
}

type CURSORINFO struct {
	CbSize      DWORD
	Flags       DWORD
	HCursor     HCURSOR
	PTScreenPos POINT
}

type ICONINFO struct {
	FIcon    BOOL
	XHotspot DWORD
	YHotspot DWORD
	HbmMask  HBITMAP
	HbmColor HBITMAP
}

const (
	CURSOR_SHOWING = 0x00000001
)
