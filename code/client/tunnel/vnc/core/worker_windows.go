package core

import (
	"fmt"
	"runtime"
	"syscall"

	"github.com/lwch/logging"
	"golang.org/x/sys/windows"
)

type ctxOsBased struct {
	hwnd   uintptr
	hdc    uintptr
	buffer uintptr
}

func (ctx *Context) attachDesktop() (func(), error) {
	runtime.LockOSThread()
	locked := true
	oldDesktop, _, err := syscall.Syscall(funcGetThreadDesktop, 1, uintptr(windows.GetCurrentThreadId()), 0, 0)
	if oldDesktop == 0 {
		return nil, fmt.Errorf("get thread desktop: %v", err)
	}
	desktop, _, err := syscall.Syscall(funcOpenInputDesktop, 3, 0, 0, windows.GENERIC_ALL)
	if desktop == 0 {
		return nil, fmt.Errorf("open input desktop: %v", err)
	}
	ok, _, err := syscall.Syscall(funcSetThreadDesktop, 1, desktop, 0, 0)
	if ok == 0 {
		logging.Error("set thread desktop: %v", err)
	}
	return func() {
		syscall.Syscall(funcSetThreadDesktop, 1, oldDesktop, 0, 0)
		syscall.Syscall(funcCloseDesktop, 1, desktop, 0, 0)
		if locked {
			runtime.UnlockOSThread()
			locked = false
		}
	}, nil
}

func (ctx *Context) init() error {
	detach, err := ctx.attachDesktop()
	if err != nil {
		return err
	}
	defer detach()
	err = ctx.getHandle()
	if err != nil {
		return err
	}
	err = ctx.updateInfo()
	if err != nil {
		return err
	}
	detach()
	return ctx.updateBuffer()
}

func (ctx *Context) getHandle() error {
	hwnd, _, err := syscall.Syscall(funcGetDesktopWindow, 0, 0, 0, 0)
	if hwnd == 0 {
		return fmt.Errorf("get desktop window: %v", err)
	}
	hdc, _, err := syscall.Syscall(funcGetDC, 1, hwnd, 0, 0)
	if hdc == 0 {
		return fmt.Errorf("get dc: %v", err)
	}
	if ctx.hdc != 0 {
		syscall.Syscall(funcReleaseDC, 2, ctx.hwnd, ctx.hdc, 0)
	}
	ctx.hwnd = hwnd
	ctx.hdc = hdc
	return nil
}

func (ctx *Context) updateInfo() error {
	bits, _, err := syscall.Syscall(funcGetDeviceCaps, 2, ctx.hdc, BITSPIXEL, 0)
	if bits == 0 {
		return fmt.Errorf("get device caps(bits): %v", err)
	}
	width, _, err := syscall.Syscall(funcGetDeviceCaps, 2, ctx.hdc, HORZRES, 0)
	if width == 0 {
		return fmt.Errorf("get device caps(width): %v", err)
	}
	height, _, err := syscall.Syscall(funcGetDeviceCaps, 2, ctx.hdc, VERTRES, 0)
	if height == 0 {
		return fmt.Errorf("get device caps(height): %v", err)
	}
	ctx.Info.Bits = int(bits)
	ctx.Info.Width = int(width)
	ctx.Info.Height = int(height)
	return nil
}

func (ctx *Context) updateBuffer() error {
	addr, _, err := syscall.Syscall(funcGlobalAlloc, 2, GMEM_FIXED, uintptr(ctx.Info.Bits*ctx.Info.Width*ctx.Info.Height/8), 0)
	if addr == 0 {
		return fmt.Errorf("global alloc: %v", err)
	}
	if ctx.buffer != 0 {
		syscall.Syscall(funcGlobalFree, 1, ctx.buffer, 0, 0)
	}
	ctx.buffer = addr
	return nil
}
