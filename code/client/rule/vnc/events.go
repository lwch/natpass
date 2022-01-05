package vnc

import (
	"encoding/json"

	"github.com/jkstack/natpass/code/client/pool"
	"github.com/lwch/logging"
)

func (v *VNC) mouseEvent(remote *pool.Conn, data []byte) {
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
	remote.SendVNCMouse(v.link.target, v.link.targetIdx, v.link.id,
		payload.Payload.Button, payload.Payload.Status, payload.Payload.X, payload.Payload.Y)
}

func (v *VNC) keyboardEvent(remote *pool.Conn, data []byte) {
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
	remote.SendVNCKeyboard(v.link.target, v.link.targetIdx, v.link.id,
		payload.Payload.Status, payload.Payload.Key)
}

func (v *VNC) cadEvent(remote *pool.Conn) {
	remote.SendVNCCADEvent(v.link.target, v.link.targetIdx, v.link.id)
}

func (v *VNC) scrollEvent(remote *pool.Conn, data []byte) {
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
	remote.SendVNCScroll(v.link.target, v.link.targetIdx, v.link.id,
		payload.Payload.X, payload.Payload.Y)
}
