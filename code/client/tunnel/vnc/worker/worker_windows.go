package worker

import (
	"fmt"
	"natpass/code/client/tunnel/vnc/define"
	"runtime"
	"syscall"

	"github.com/lwch/logging"
	"golang.org/x/sys/windows"
)

type workerOsBased struct {
	buffer uintptr
}

func attachDesktop() (func(), error) {
	runtime.LockOSThread()
	locked := true
	oldDesktop, _, err := syscall.Syscall(define.FuncGetThreadDesktop, 1, uintptr(windows.GetCurrentThreadId()), 0, 0)
	if oldDesktop == 0 {
		runtime.UnlockOSThread()
		return nil, fmt.Errorf("get thread desktop: %v", err)
	}
	desktop, _, err := syscall.Syscall(define.FuncOpenInputDesktop, 3, 0, 0, windows.GENERIC_ALL)
	if desktop == 0 {
		runtime.UnlockOSThread()
		return nil, fmt.Errorf("open input desktop: %v", err)
	}
	ok, _, err := syscall.Syscall(define.FuncSetThreadDesktop, 1, desktop, 0, 0)
	if ok == 0 {
		syscall.Syscall(define.FuncCloseDesktop, 1, desktop, 0, 0)
		runtime.UnlockOSThread()
		return nil, fmt.Errorf("set thread desktop: %v", err)
	}
	return func() {
		syscall.Syscall(define.FuncSetThreadDesktop, 1, oldDesktop, 0, 0)
		syscall.Syscall(define.FuncCloseDesktop, 1, desktop, 0, 0)
		if locked {
			runtime.UnlockOSThread()
			locked = false
		}
	}, nil
}

func (worker *Worker) init() error {
	detach, err := attachDesktop()
	if err != nil {
		return err
	}
	defer detach()
	cancel, hdc, err := worker.getHandle()
	if err != nil {
		return err
	}
	defer cancel()
	err = worker.updateInfo(hdc)
	if err != nil {
		return err
	}
	detach()
	return worker.updateBuffer()
}

func (worker *workerOsBased) getHandle() (func(), uintptr, error) {
	hwnd, _, err := syscall.Syscall(define.FuncGetDesktopWindow, 0, 0, 0, 0)
	if hwnd == 0 {
		return nil, 0, fmt.Errorf("get desktop window: %v", err)
	}
	hdc, _, err := syscall.Syscall(define.FuncGetDC, 1, hwnd, 0, 0)
	if hdc == 0 {
		return nil, 0, fmt.Errorf("get dc: %v", err)
	}
	return func() {
		syscall.Syscall(define.FuncReleaseDC, 2, hwnd, hdc, 0)
	}, hdc, nil
}

func (worker *Worker) updateInfo(hdc uintptr) error {
	bits, _, err := syscall.Syscall(define.FuncGetDeviceCaps, 2, hdc, define.BITSPIXEL, 0)
	if bits == 0 {
		return fmt.Errorf("get device caps(bits): %v", err)
	}
	width, _, err := syscall.Syscall(define.FuncGetDeviceCaps, 2, hdc, define.HORZRES, 0)
	if width == 0 {
		return fmt.Errorf("get device caps(width): %v", err)
	}
	height, _, err := syscall.Syscall(define.FuncGetDeviceCaps, 2, hdc, define.VERTRES, 0)
	if height == 0 {
		return fmt.Errorf("get device caps(height): %v", err)
	}
	worker.info.bits = int(bits)
	worker.info.width = int(width)
	worker.info.height = int(height)
	if worker.info.bits != 32 {
		worker.info.bits = 32
		logging.Info("reset bits to 32")
	}
	return nil
}

func (worker *Worker) updateBuffer() error {
	addr, _, err := syscall.Syscall(define.FuncGlobalAlloc, 2, define.GMEM_FIXED, uintptr(worker.info.bits*worker.info.width*worker.info.height/8), 0)
	if addr == 0 {
		return fmt.Errorf("global alloc: %v", err)
	}
	if worker.buffer != 0 {
		syscall.Syscall(define.FuncGlobalFree, 1, worker.buffer, 0, 0)
	}
	worker.buffer = addr
	return nil
}
