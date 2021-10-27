package core

import (
	"errors"
	"natpass/code/client/tunnel/vnc/core/define"
	"syscall"
	"unsafe"

	"github.com/gorilla/websocket"
	"github.com/lwch/logging"
)

func (worker *Worker) runCapture(conn *websocket.Conn) {
	err := worker.capture()
	if err != nil {
		logging.Error(err.Error())
		captureError(conn, err.Error())
		return
	}
	data := make([]byte, worker.info.width*worker.info.height*worker.info.bits/8)
	for i := 0; i < len(data); i++ {
		data[i] = *(*uint8)(unsafe.Pointer(worker.buffer + uintptr(i)))
	}
	captureOK(conn, worker.info.bits, worker.info.width, worker.info.height, data)
}

func (worker *Worker) capture() error {
	detach, err := attachDesktop()
	if err != nil {
		return errors.New("attach desktop: " + err.Error())
	}
	defer detach()
	info := worker.info
	err = worker.updateInfo()
	if err != nil {
		return errors.New("update info: " + err.Error())
	}
	if info.bits != worker.info.bits ||
		info.width != worker.info.width ||
		info.height != worker.info.height {
		err = worker.updateBuffer()
		if err != nil {
			return errors.New("update buffer: " + err.Error())
		}
	}
	memDC, _, err := syscall.Syscall(define.FuncCreateCompatibleDC, 1, worker.hdc, 0, 0)
	if memDC == 0 {
		return errors.New("create dc: " + err.Error())
	}
	defer syscall.Syscall(define.FuncDeleteDC, 1, memDC, 0, 0)
	bitmap, _, err := syscall.Syscall(define.FuncCreateCompatibleBitmap, 3, worker.hdc,
		uintptr(worker.info.width), uintptr(worker.info.height))
	if bitmap == 0 {
		return errors.New("create bitmap: " + err.Error())
	}
	defer syscall.Syscall(define.FuncDeleteObject, 1, bitmap, 0, 0)
	oldDC, _, err := syscall.Syscall(define.FuncSelectObject, 2, memDC, bitmap, 0)
	if oldDC == 0 {
		return errors.New("select object: " + err.Error())
	}
	defer syscall.Syscall(define.FuncSelectObject, 2, memDC, oldDC, 0)
	ok, _, err := syscall.Syscall9(define.FuncBitBlt, 0, memDC, 0, 0,
		uintptr(worker.info.width), uintptr(worker.info.height), worker.hdc, 0, 0, define.SRCCOPY)
	if ok == 0 {
		return errors.New("bitblt: " + err.Error())
	}
	defer worker.copyImageData(bitmap)
	// TODO: draw cursor
	return nil
}

type BITMAPINFOHEADER struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

func (worker *Worker) copyImageData(bitmap uintptr) {
	var hdr BITMAPINFOHEADER
	hdr.BiSize = uint32(unsafe.Sizeof(hdr))
	hdr.BiPlanes = 1
	hdr.BiBitCount = uint16(worker.info.bits)
	hdr.BiWidth = int32(worker.info.width)
	hdr.BiHeight = int32(-worker.info.height)
	hdr.BiCompression = define.BI_RGB
	hdr.BiSizeImage = 0
	syscall.Syscall9(define.FuncGetDIBits, 7, worker.hdc, bitmap, 0, uintptr(worker.info.height),
		worker.buffer, uintptr(unsafe.Pointer(&hdr)), define.DIB_RGB_COLORS, 0, 0)
}
