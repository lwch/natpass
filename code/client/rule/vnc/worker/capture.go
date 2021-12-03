package worker

import (
	"fmt"
	"natpass/code/client/rule/vnc/define"
	"natpass/code/client/rule/vnc/vncnetwork"
	"runtime"
	"syscall"

	"github.com/lwch/logging"
	"golang.org/x/sys/windows"
)

func (worker *Worker) runCapture() vncnetwork.ImageData {
	detach, err := attachDesktop()
	if err != nil {
		logging.Error("attach desktop: " + err.Error())
		return vncnetwork.ImageData{
			Ok:  false,
			Msg: fmt.Sprintf("attach desktop: " + err.Error()),
		}
	}
	defer detach()
	img, err := worker.cli.Screenshot()
	if err != nil {
		logging.Error("screenshot: " + err.Error())
		return vncnetwork.ImageData{
			Ok:  false,
			Msg: fmt.Sprintf("screenshot: " + err.Error()),
		}
	}
	data := make([]byte, len(img.Pix))
	copy(data, img.Pix)
	return vncnetwork.ImageData{
		Ok:     true,
		Bits:   32,
		Width:  uint32(img.Rect.Max.X),
		Height: uint32(img.Rect.Max.Y),
		Data:   data,
	}
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
