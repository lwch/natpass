package core

import "syscall"

var (
	libUser32, _            = syscall.LoadLibrary("user32.dll")
	funcOpenInputDesktop, _ = syscall.GetProcAddress(syscall.Handle(libUser32), "OpenInputDesktop")
	funcSetThreadDesktop, _ = syscall.GetProcAddress(syscall.Handle(libUser32), "SetThreadDesktop")
	funcGetThreadDesktop, _ = syscall.GetProcAddress(syscall.Handle(libUser32), "GetThreadDesktop")
	funcCloseDesktop, _     = syscall.GetProcAddress(syscall.Handle(libUser32), "CloseDesktop")
	funcGetDesktopWindow, _ = syscall.GetProcAddress(syscall.Handle(libUser32), "GetDesktopWindow")
	funcGetDC, _            = syscall.GetProcAddress(syscall.Handle(libUser32), "GetDC")
	funcReleaseDC, _        = syscall.GetProcAddress(syscall.Handle(libUser32), "ReleaseDC")
)
