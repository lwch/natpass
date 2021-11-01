package worker

import (
	"natpass/code/client/tunnel/vnc/vncnetwork"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

func captureError(conn *websocket.Conn, msg string) {
	var m vncnetwork.VncMsg
	m.XType = vncnetwork.VncMsg_capture_data
	m.Payload = &vncnetwork.VncMsg_Data{
		Data: &vncnetwork.ImageData{
			Ok:  false,
			Msg: msg,
		},
	}
	data, _ := proto.Marshal(&m)
	conn.WriteMessage(websocket.BinaryMessage, data)
}

func captureOK(conn *websocket.Conn, bits, width, height int, data []byte) {
	var msg vncnetwork.VncMsg
	msg.XType = vncnetwork.VncMsg_capture_data
	msg.Payload = &vncnetwork.VncMsg_Data{
		Data: &vncnetwork.ImageData{
			Ok:     true,
			Bits:   uint32(bits),
			Width:  uint32(width),
			Height: uint32(height),
			Data:   data,
		},
	}
	data, _ = proto.Marshal(&msg)
	conn.WriteMessage(websocket.BinaryMessage, data)
}
