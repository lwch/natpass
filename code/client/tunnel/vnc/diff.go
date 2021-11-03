package vnc

import (
	"image"
	"reflect"
	"unsafe"
)

func calcDiff(src, dst *image.RGBA) []image.Rectangle {
	// 宽度必须为2的倍数
	const width = 64
	const height = 64
	size := dst.Bounds()
	ret := make([]image.Rectangle, 0, (size.Max.X*size.Max.Y)/(width*height))
	for y := 0; y < size.Max.Y; y += height {
		for x := 0; x < size.Max.X; x += width {
			dWidth := size.Max.X - x
			dHeight := size.Max.Y - y
			if dWidth > width {
				dWidth = width
			}
			if dHeight > height {
				dHeight = height
			}
			rect := image.Rect(x, y, x+dWidth, y+dHeight)
			if dWidth%2 == 0 {
				if isDiff8(src, dst, rect) {
					ret = append(ret, rect)
				}
			} else {
				if isDiff4(src, dst, rect) {
					ret = append(ret, rect)
				}
			}
			////////
		}
	}
	return ret
}

func isDiff8(src, dst *image.RGBA, rect image.Rectangle) bool {
	sx := src.Bounds().Max.X * 4
	dx := rect.Min.X * 4
	ptr := uintptr(rect.Min.Y*sx + dx)
	srcData := unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&src.Pix)).Data)
	dstData := unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&dst.Pix)).Data)
	for y := 0; y < rect.Size().Y; y++ {
		next := ptr + uintptr(sx)
		for x := 0; x < rect.Size().X; x += 2 {
			src := (*uint64)(unsafe.Pointer(uintptr(srcData) + ptr))
			dst := (*uint64)(unsafe.Pointer(uintptr(dstData) + ptr))
			if *src != *dst {
				return true
			}
			ptr += 8
		}
		ptr = next
	}
	return false
}

func isDiff4(src, dst *image.RGBA, rect image.Rectangle) bool {
	sx := src.Bounds().Max.X * 4
	dx := rect.Min.X * 4
	ptr := uintptr(rect.Min.Y*sx + dx)
	srcData := unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&src.Pix)).Data)
	dstData := unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&dst.Pix)).Data)
	for y := 0; y < rect.Size().Y; y++ {
		next := ptr + uintptr(sx)
		for x := 0; x < rect.Size().X; x++ {
			src := (*uint64)(unsafe.Pointer(uintptr(srcData) + ptr))
			dst := (*uint64)(unsafe.Pointer(uintptr(dstData) + ptr))
			if *src != *dst {
				return true
			}
			ptr += 4
		}
		ptr = next
	}
	return false
}
