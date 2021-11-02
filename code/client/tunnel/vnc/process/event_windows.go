package process

import (
	"natpass/code/client/tunnel/vnc/vncnetwork"
	"natpass/code/network"
)

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
