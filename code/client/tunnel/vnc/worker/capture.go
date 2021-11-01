package worker

import (
	"natpass/code/client/tunnel/vnc/network"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

func captureError(conn *websocket.Conn, msg string) {
	var m network.VncMsg
	m.XType = network.VncMsg_capture_data
	m.Payload = &network.VncMsg_Data{
		Data: &network.ImageData{
			Ok:  false,
			Msg: msg,
		},
	}
	data, _ := proto.Marshal(&m)
	conn.WriteMessage(websocket.BinaryMessage, data)
}

func captureOK(conn *websocket.Conn, bits, width, height int, data []byte) {
	var msg network.VncMsg
	msg.XType = network.VncMsg_capture_data
	msg.Payload = &network.VncMsg_Data{
		Data: &network.ImageData{
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
