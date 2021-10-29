package core

import (
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

func captureError(conn *websocket.Conn, msg string) {
	var m VncMsg
	m.XType = VncMsg_capture_data
	m.Payload = &VncMsg_Data{
		Data: &ImageData{
			Ok:  false,
			Msg: msg,
		},
	}
	data, _ := proto.Marshal(&m)
	conn.WriteMessage(websocket.BinaryMessage, data)
}

func captureOK(conn *websocket.Conn, bits, width, height int, data []byte) {
	var msg VncMsg
	msg.XType = VncMsg_capture_data
	msg.Payload = &VncMsg_Data{
		Data: &ImageData{
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
