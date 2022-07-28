package vnc

import (
	"bytes"
	"image"
	"image/jpeg"
	"time"

	"github.com/lwch/logging"
	"github.com/lwch/natpass/code/client/conn"
	"github.com/lwch/natpass/code/client/rule/vnc/process"
	"github.com/lwch/natpass/code/network"
	"github.com/lwch/natpass/code/utils"
)

const (
	zoneWidth  = 64
	zoneHeight = 64
)

// Link vnc link
type Link struct {
	parent *VNC
	id     string // link id
	target string // target id
	remote *conn.Conn
	// vnc
	ps         *process.Process
	quality    uint32
	showCursor bool
	img        *image.RGBA
	// runtime
	sendBytes  uint64
	recvBytes  uint64
	sendPacket uint64
	recvPacket uint64
	idx        int
	reDraw     bool
}

// GetID get link id
func (link *Link) GetID() string {
	return link.id
}

// GetBytes get send and recv bytes
func (link *Link) GetBytes() (uint64, uint64) {
	return link.recvBytes, link.sendBytes
}

// GetPackets get send and recv packets
func (link *Link) GetPackets() (uint64, uint64) {
	return link.recvPacket, link.sendPacket
}

// SetQuality transfer quality
func (link *Link) SetQuality(q uint32) {
	link.quality = q
	link.reDraw = true
}

// SetCursor set show cursor
func (link *Link) SetCursor(b bool) {
	link.showCursor = b
	link.reDraw = true
}

// Fork fork worker process
func (link *Link) Fork(confDir string) error {
	p, err := process.CreateWorker(link.parent.Name, confDir, link.showCursor)
	if err != nil {
		return err
	}
	link.ps = p
	return nil
}

// Forward forward data
func (link *Link) Forward() {
	go link.remoteRead()
	go link.localRead()
}

func (link *Link) remoteRead() {
	defer link.Close(true)
	ch := link.remote.ChanRead(link.id)
	for {
		msg := <-ch
		if msg == nil {
			return
		}
		switch msg.GetXType() {
		case network.Msg_vnc_ctrl:
			ctrl := msg.GetVctrl()
			link.SetQuality(ctrl.GetQuality())
			link.SetCursor(ctrl.GetCursor())
			link.ps.SetCursor(ctrl.GetCursor())
		case network.Msg_vnc_mouse:
			link.ps.MouseEvent(msg.GetVmouse())
		case network.Msg_vnc_keyboard:
			link.ps.KeyboardEvent(msg.GetVkbd())
		case network.Msg_vnc_cad:
			link.ps.CADEvent()
		case network.Msg_vnc_scroll:
			link.ps.ScrollEvent(msg.GetVscroll())
		case network.Msg_vnc_clipboard:
			if msg.GetVclipboard().GetSet() {
				link.ps.SetClipboard(msg.GetVclipboard())
			} else {
				data := link.ps.GetClipboard()
				link.remote.SendVNCClipboardData(link.target, link.id, true, data)
			}
		}
	}
}

func (link *Link) localRead() {
	// TODO: exit by context
	defer utils.Recover("capture")
	defer link.Close(true)
	img, err := link.ps.Capture(3 * time.Second)
	if err != nil {
		logging.Error("capture: %v", err)
		return
	}
	link.sendAll(img)
	link.img = img
	size := img.Rect
	sleep := time.Second / time.Duration(link.parent.cfg.Fps)
	for {
		time.Sleep(sleep)
		img, err = link.ps.Capture(0)
		if err != nil {
			logging.Error("capture: %v", err)
			continue
		}
		if img.Rect.Dx() != size.Dx() ||
			img.Rect.Dy() != size.Dy() ||
			link.reDraw ||
			link.idx%10000 == 0 {
			link.sendAll(img)
			link.reDraw = false
		} else {
			link.sendDiff(img)
		}
		link.img = img
		size = img.Rect
		link.idx++
	}
}

// Close close link
func (link *Link) Close(send bool) {
	if link.ps != nil {
		link.ps.Close()
	}
	if send {
		link.remote.SendDisconnect(link.target, link.id)
	}
	link.parent.remove(link.id)
	link.remote.ChanClose(link.id)
}

func cut(src *image.RGBA, rect image.Rectangle) *image.RGBA {
	size := rect.Size()
	ret := image.NewRGBA(image.Rect(0, 0, size.X, size.Y))
	sx := src.Bounds().Size().X * 4
	dx := rect.Min.X * 4
	idx := rect.Min.Y*sx + dx
	for y := 0; y < size.Y; y++ {
		copy(ret.Pix[y*size.X*4:], src.Pix[idx:idx+size.X*4-1])
		idx += sx
	}
	return ret
}

func (link *Link) sendAll(img *image.RGBA) {
	size := img.Bounds()
	screen := image.Rect(0, 0, img.Rect.Dx(), img.Rect.Dy())
	var buf bytes.Buffer
	for y := 0; y < size.Max.Y; y += zoneHeight {
		for x := 0; x < size.Max.X; x += zoneWidth {
			width := size.Max.X - x
			height := size.Max.Y - y
			if width > zoneWidth {
				width = zoneWidth
			}
			if height > zoneHeight {
				height = zoneHeight
			}
			rect := image.Rect(x, y, x+width, y+height)
			next := cut(img, rect)
			if link.quality == 100 {
				link.remote.SendVNCImage(link.target, link.id,
					screen, rect, network.VncImage_raw, next.Pix)
				continue
			}
			buf.Reset()
			err := jpeg.Encode(&buf, next, &jpeg.Options{Quality: int(link.quality)})
			if err == nil {
				link.remote.SendVNCImage(link.target, link.id,
					screen, rect, network.VncImage_jpeg, buf.Bytes())
			} else {
				link.remote.SendVNCImage(link.target, link.id,
					screen, rect, network.VncImage_raw, next.Pix)
			}
		}
	}
}

func (link *Link) sendDiff(img *image.RGBA) {
	blocks := calcDiff(link.img, img)
	screen := image.Rect(0, 0, img.Rect.Dx(), img.Rect.Dy())
	var buf bytes.Buffer
	for _, block := range blocks {
		next := cut(img, block)
		if link.quality == 100 {
			link.remote.SendVNCImage(link.target, link.id,
				screen, block, network.VncImage_raw, next.Pix)
			continue
		}
		buf.Reset()
		err := jpeg.Encode(&buf, next, &jpeg.Options{Quality: int(link.quality)})
		if err == nil {
			link.remote.SendVNCImage(link.target, link.id,
				screen, block, network.VncImage_jpeg, buf.Bytes())
		} else {
			link.remote.SendVNCImage(link.target, link.id,
				screen, block, network.VncImage_raw, next.Pix)
		}
	}
}
