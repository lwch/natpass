package vnc

import (
	"encoding/json"
	"natpass/code/client/pool"

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
