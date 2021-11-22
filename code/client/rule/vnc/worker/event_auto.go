//go:build vnc
// +build vnc

package worker

import (
	"natpass/code/client/rule/vnc/vncnetwork"

	"github.com/go-vgo/robotgo"
	"github.com/gorilla/websocket"
	"github.com/lwch/logging"
	"google.golang.org/protobuf/proto"
)

func runMouse(data *vncnetwork.MouseData) {
	detach, err := attachDesktop()
	if err != nil {
		logging.Error("attach desktop: %v", err)
		return
	}
	defer detach()
	robotgo.MoveMouse(int(data.GetX()), int(data.GetY()))
	var key string
	switch data.GetBtn() {
	case vncnetwork.MouseData_left:
		key = "left"
	case vncnetwork.MouseData_right:
		key = "right"
	case vncnetwork.MouseData_middle:
		key = "center"
	}
	switch data.GetType() {
	case vncnetwork.Status_down:
		robotgo.MouseToggle("down", key)
	case vncnetwork.Status_up:
		robotgo.MouseToggle("up", key)
	}
}

func runKeyboard(data *vncnetwork.KeyboardData) {
	detach, err := attachDesktop()
	if err != nil {
		logging.Error("attach desktop: %v", err)
		return
	}
	defer detach()
	switch data.Type {
	case vncnetwork.Status_down:
		robotgo.KeyToggle(data.Key, "down")
	case vncnetwork.Status_up:
		robotgo.KeyToggle(data.Key, "up")
	}
}

func runScroll(data *vncnetwork.ScrollData) {
	detach, err := attachDesktop()
	if err != nil {
		logging.Error("attach desktop: %v", err)
		return
	}
	defer detach()
	robotgo.Scroll(int(data.X), int(data.Y), 1)
}

func runClipboard(conn *websocket.Conn, data *vncnetwork.ClipboardData) {
	detach, err := attachDesktop()
	if err != nil {
		logging.Error("attach desktop: %v", err)
		return
	}
	defer detach()
	if data.GetSet() {
		setClipboard(data)
	} else {
		getClipboard(conn)
	}
}

func setClipboard(data *vncnetwork.ClipboardData) {
	switch data.GetXType() {
	case vncnetwork.ClipboardData_file:
	case vncnetwork.ClipboardData_image:
	case vncnetwork.ClipboardData_text:
		robotgo.WriteAll(data.GetData())
	}
}

func getClipboard(conn *websocket.Conn) {
	data, _ := robotgo.ReadAll()
	var msg vncnetwork.VncMsg
	msg.XType = vncnetwork.VncMsg_clipboard_event
	msg.Payload = &vncnetwork.VncMsg_Clipboard{
		Clipboard: &vncnetwork.ClipboardData{
			Set:   true,
			XType: vncnetwork.ClipboardData_text,
			Payload: &vncnetwork.ClipboardData_Data{
				Data: data,
			},
		},
	}
	enc, _ := proto.Marshal(&msg)
	conn.WriteMessage(websocket.BinaryMessage, enc)
}
