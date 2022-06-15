package vnc

import (
	"encoding/json"

	"github.com/lwch/logging"
	"github.com/lwch/natpass/code/client/conn"
)

func (v *VNC) mouseEvent(remote *conn.Conn, data []byte) {
	var payload struct {
		Payload struct {
			Button string `json:"button"`
			Status string `json:"status"`
			X      int    `json:"x"`
			Y      int    `json:"y"`
		} `json:"payload"`
	}
	err := json.Unmarshal(data, &payload)
	if err != nil {
		logging.Error("unmarshal: %v", err)
		return
	}
	remote.SendVNCMouse(v.link.target, v.link.id,
		payload.Payload.Button, payload.Payload.Status, payload.Payload.X, payload.Payload.Y)
}

func (v *VNC) keyboardEvent(remote *conn.Conn, data []byte) {
	var payload struct {
		Payload struct {
			Status string `json:"status"`
			Key    string `json:"key"`
		} `json:"payload"`
	}
	err := json.Unmarshal(data, &payload)
	if err != nil {
		logging.Error("unmarshal: %v", err)
		return
	}
	remote.SendVNCKeyboard(v.link.target, v.link.id,
		payload.Payload.Status, payload.Payload.Key)
}

func (v *VNC) cadEvent(remote *conn.Conn) {
	remote.SendVNCCADEvent(v.link.target, v.link.id)
}

func (v *VNC) scrollEvent(remote *conn.Conn, data []byte) {
	var payload struct {
		Payload struct {
			X int32 `json:"x"`
			Y int32 `json:"y"`
		} `json:"payload"`
	}
	err := json.Unmarshal(data, &payload)
	if err != nil {
		logging.Error("unmarshal: %v", err)
		return
	}
	remote.SendVNCScroll(v.link.target, v.link.id,
		payload.Payload.X, payload.Payload.Y)
}
