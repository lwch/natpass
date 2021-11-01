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
)
