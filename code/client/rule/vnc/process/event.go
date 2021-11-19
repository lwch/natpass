package process

import (
	"natpass/code/client/rule/vnc/vncnetwork"
	"natpass/code/network"
)

// MouseEvent dispatch mouse event to child process
func (p *Process) MouseEvent(data *network.VncMouse) {
	t := vncnetwork.Status_unset_st
	switch data.GetType() {
	case network.VncStatus_down:
		t = vncnetwork.Status_down
	case network.VncStatus_up:
		t = vncnetwork.Status_up
	}
	btn := vncnetwork.MouseData_unset_btn
	switch data.GetBtn() {
	case network.VncMouse_left:
		btn = vncnetwork.MouseData_left
	case network.VncMouse_middle:
		btn = vncnetwork.MouseData_middle
	case network.VncMouse_right:
		btn = vncnetwork.MouseData_right
	}
	var msg vncnetwork.VncMsg
	msg.XType = vncnetwork.VncMsg_mouse_event
	msg.Payload = &vncnetwork.VncMsg_Mouse{
		Mouse: &vncnetwork.MouseData{
			Type: t,
			Btn:  btn,
			X:    data.GetX(),
			Y:    data.GetY(),
		},
	}
	p.chWrite <- &msg
}

// KeyboardEvent dispatch keyboard event to child process
func (p *Process) KeyboardEvent(data *network.VncKeyboard) {
	t := vncnetwork.Status_unset_st
	switch data.GetType() {
	case network.VncStatus_down:
		t = vncnetwork.Status_down
	case network.VncStatus_up:
		t = vncnetwork.Status_up
	}
	var msg vncnetwork.VncMsg
	msg.XType = vncnetwork.VncMsg_keyboard_event
	msg.Payload = &vncnetwork.VncMsg_Keyboard{
		Keyboard: &vncnetwork.KeyboardData{
			Type: t,
			Key:  data.GetKey(),
		},
	}
	p.chWrite <- &msg
}

// SetCursor dispatch draw cursor to child process
func (p *Process) SetCursor(b bool) {
	var msg vncnetwork.VncMsg
	msg.XType = vncnetwork.VncMsg_set_cursor
	msg.Payload = &vncnetwork.VncMsg_ShowCursor{
		ShowCursor: b,
	}
	p.chWrite <- &msg
}

// ScrollEvent dispatch scroll event to child process
func (p *Process) ScrollEvent(data *network.VncScroll) {
	var msg vncnetwork.VncMsg
	msg.XType = vncnetwork.VncMsg_scroll_event
	msg.Payload = &vncnetwork.VncMsg_Scroll{
		Scroll: &vncnetwork.ScrollData{
			X: data.GetX(),
			Y: data.GetY(),
		},
	}
	p.chWrite <- &msg
}
