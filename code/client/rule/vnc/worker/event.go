package worker

import (
	"github.com/gorilla/websocket"
	"github.com/lwch/logging"
	"github.com/lwch/natpass/code/client/rule/vnc/vncnetwork"
	"github.com/lwch/rdesktop"
	"google.golang.org/protobuf/proto"
)

func (worker *Worker) runMouse(data *vncnetwork.MouseData) {
	detach, err := attachDesktop()
	if err != nil {
		logging.Error("attach desktop: %v", err)
		return
	}
	defer detach()
	worker.cli.MouseMove(int(data.GetX()), int(data.GetY()))
	var button rdesktop.MouseButton
	switch data.GetBtn() {
	case vncnetwork.MouseData_left:
		button = rdesktop.MouseLeft
	case vncnetwork.MouseData_right:
		button = rdesktop.MouseRight
	case vncnetwork.MouseData_middle:
		button = rdesktop.MouseMiddle
	}
	switch data.GetType() {
	case vncnetwork.Status_down:
		worker.cli.ToggleMouse(button, true)
	case vncnetwork.Status_up:
		worker.cli.ToggleMouse(button, false)
	}
}

func (worker *Worker) runKeyboard(data *vncnetwork.KeyboardData) {
	detach, err := attachDesktop()
	if err != nil {
		logging.Error("attach desktop: %v", err)
		return
	}
	defer detach()
	switch data.Type {
	case vncnetwork.Status_down:
		worker.cli.ToggleKey(data.Key, true)
	case vncnetwork.Status_up:
		worker.cli.ToggleKey(data.Key, false)
	}
}

func (worker *Worker) runScroll(data *vncnetwork.ScrollData) {
	detach, err := attachDesktop()
	if err != nil {
		logging.Error("attach desktop: %v", err)
		return
	}
	defer detach()
	worker.cli.Scroll(int(data.X), int(data.Y))
}

func (worker *Worker) runClipboard(conn *websocket.Conn, data *vncnetwork.ClipboardData) {
	detach, err := attachDesktop()
	if err != nil {
		logging.Error("attach desktop: %v", err)
		return
	}
	defer detach()
	if data.GetSet() {
		worker.setClipboard(data)
	} else {
		worker.getClipboard(conn)
	}
}

func (worker *Worker) setClipboard(data *vncnetwork.ClipboardData) {
	switch data.GetXType() {
	case vncnetwork.ClipboardData_file:
	case vncnetwork.ClipboardData_image:
	case vncnetwork.ClipboardData_text:
		worker.cli.ClipboardSet(data.GetData())
	}
}

func (worker *Worker) getClipboard(conn *websocket.Conn) {
	data, _ := worker.cli.ClipboardGet()
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
