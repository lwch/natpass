package pool

import (
	"image"
	"natpass/code/network"
	"time"
)

// SendVNCImage send vnc image data
func (conn *Conn) SendVNCImage(to string, toIdx uint32, id string, screen, rect image.Rectangle,
	encode network.VncImageEncoding, data []byte) {
	var msg network.Msg
	msg.To = to
	msg.ToIdx = toIdx
	msg.XType = network.Msg_vnc_image
	msg.LinkId = id
	msg.Payload = &network.Msg_Vimg{
		Vimg: &network.VncImage{
			XInfo: &network.VncImageInfo{
				ScreenWidth:  uint32(screen.Dx()),
				ScreenHeight: uint32(screen.Dy()),
				RectX:        uint32(rect.Min.X),
				RectY:        uint32(rect.Min.Y),
				RectWidth:    uint32(rect.Dx()),
				RectHeight:   uint32(rect.Dy()),
			},
			Encode: encode,
			Data:   data,
		},
	}
	select {
	case conn.write <- &msg:
	case <-time.After(conn.parent.cfg.WriteTimeout):
	}
}

// SendVNCCtrl send vnc config
func (conn *Conn) SendVNCCtrl(to string, toIdx uint32, id string, quality uint64, showCursor bool) {
	var msg network.Msg
	msg.To = to
	msg.ToIdx = toIdx
	msg.XType = network.Msg_vnc_ctrl
	msg.LinkId = id
	msg.Payload = &network.Msg_Vctrl{
		Vctrl: &network.VncControl{
			Quality: uint32(quality),
			Cursor:  showCursor,
		},
	}
	select {
	case conn.write <- &msg:
	case <-time.After(conn.parent.cfg.WriteTimeout):
	}
}

// SendVNCMouse send vnc mouse event
func (conn *Conn) SendVNCMouse(to string, toIdx uint32, id string,
	button, status string, x, y int) {
	t := network.VncStatus_unset_st
	switch status {
	case "down":
		t = network.VncStatus_down
	case "up":
		t = network.VncStatus_up
	}
	btn := network.VncMouse_unset_btn
	switch button {
	case "left":
		btn = network.VncMouse_left
	case "middle":
		btn = network.VncMouse_middle
	case "right":
		btn = network.VncMouse_right
	}
	var msg network.Msg
	msg.To = to
	msg.ToIdx = toIdx
	msg.XType = network.Msg_vnc_mouse
	msg.LinkId = id
	msg.Payload = &network.Msg_Vmouse{
		Vmouse: &network.VncMouse{
			Type: t,
			Btn:  btn,
			X:    uint32(x),
			Y:    uint32(y),
		},
	}
	select {
	case conn.write <- &msg:
	case <-time.After(conn.parent.cfg.WriteTimeout):
	}
}

// SendVNCKeyboard send vnc keyboard event
func (conn *Conn) SendVNCKeyboard(to string, toIdx uint32, id string,
	status, key string) {
	t := network.VncStatus_unset_st
	switch status {
	case "down":
		t = network.VncStatus_down
	case "up":
		t = network.VncStatus_up
	}
	var msg network.Msg
	msg.To = to
	msg.ToIdx = toIdx
	msg.XType = network.Msg_vnc_keyboard
	msg.LinkId = id
	msg.Payload = &network.Msg_Vkbd{
		Vkbd: &network.VncKeyboard{
			Type: t,
			Key:  key,
		},
	}
	select {
	case conn.write <- &msg:
	case <-time.After(conn.parent.cfg.WriteTimeout):
	}
}
