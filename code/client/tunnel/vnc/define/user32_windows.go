package define

import "syscall"

var (
	libUser32, _ = syscall.LoadLibrary("user32.dll")
	// FuncOpenInputDesktop https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-openinputdesktop
	FuncOpenInputDesktop, _ = syscall.GetProcAddress(syscall.Handle(libUser32), "OpenInputDesktop")
	// FuncSetThreadDesktop https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-setthreaddesktop
	FuncSetThreadDesktop, _ = syscall.GetProcAddress(syscall.Handle(libUser32), "SetThreadDesktop")
	// FuncGetThreadDesktop https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getthreaddesktop
	FuncGetThreadDesktop, _ = syscall.GetProcAddress(syscall.Handle(libUser32), "GetThreadDesktop")
	// FuncCloseDesktop https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-closedesktop
	FuncCloseDesktop, _ = syscall.GetProcAddress(syscall.Handle(libUser32), "CloseDesktop")
	// FuncGetDesktopWindow https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getdesktopwindow
	FuncGetDesktopWindow, _ = syscall.GetProcAddress(syscall.Handle(libUser32), "GetDesktopWindow")
	// FuncGetDC https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getdc
	FuncGetDC, _ = syscall.GetProcAddress(syscall.Handle(libUser32), "GetDC")
	// FuncReleaseDC https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-releasedc
	FuncReleaseDC, _ = syscall.GetProcAddress(syscall.Handle(libUser32), "ReleaseDC")
	// FuncGetCursorInfo https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getcursorinfo
	FuncGetCursorInfo, _ = syscall.GetProcAddress(syscall.Handle(libUser32), "GetCursorInfo")
	// FuncGetIconInfo https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-geticoninfo
	FuncGetIconInfo, _ = syscall.GetProcAddress(syscall.Handle(libUser32), "GetIconInfo")
	// FuncDrawIcon https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-drawicon
	FuncDrawIcon, _ = syscall.GetProcAddress(syscall.Handle(libUser32), "DrawIcon")
)

type (
	// HANDLE handle object
	HANDLE uintptr
	// BOOL bool
	BOOL int32
	// DWORD double word
	DWORD uint32
	// LONG long
	LONG int32
	// HCURSOR cursor handle
	HCURSOR HANDLE
	// HBITMAP bitmap handle
	HBITMAP HANDLE
)

// POINT pointer
type POINT struct {
	X LONG
	Y LONG
}

// CURSORINFO cursor info
type CURSORINFO struct {
	CbSize      DWORD
	Flags       DWORD
	HCursor     HCURSOR
	PTScreenPos POINT
}

// ICONINFO icon info
type ICONINFO struct {
	FIcon    BOOL
	XHotspot DWORD
	YHotspot DWORD
	HbmMask  HBITMAP
	HbmColor HBITMAP
}

const (
	// CURSOR_SHOWING https://docs.microsoft.com/en-us/windows/win32/api/winuser/ns-winuser-cursorinfo
	CURSOR_SHOWING = 0x00000001
)
