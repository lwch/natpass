package worker

import (
	"fmt"
	"runtime"
	"syscall"

	"github.com/lwch/natpass/code/client/rule/vnc/define"
	"golang.org/x/sys/windows"
)

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
