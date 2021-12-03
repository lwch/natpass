package define

import "syscall"

var (
	libUser32, _ = syscall.LoadLibrary("user32.dll")
	// FuncGetThreadDesktop https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getthreaddesktop
	FuncGetThreadDesktop, _ = syscall.GetProcAddress(syscall.Handle(libUser32), "GetThreadDesktop")
	// FuncOpenInputDesktop https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-openinputdesktop
	FuncOpenInputDesktop, _ = syscall.GetProcAddress(syscall.Handle(libUser32), "OpenInputDesktop")
	// FuncSetThreadDesktop https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-setthreaddesktop
	FuncSetThreadDesktop, _ = syscall.GetProcAddress(syscall.Handle(libUser32), "SetThreadDesktop")
	// FuncCloseDesktop https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-closedesktop
	FuncCloseDesktop, _ = syscall.GetProcAddress(syscall.Handle(libUser32), "CloseDesktop")
)
