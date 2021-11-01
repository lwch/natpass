package worker

import (
	"image"
	"image/jpeg"
	"natpass/code/client/tunnel/vnc/vncnetwork"
	"os"

	"github.com/gorilla/websocket"
	"github.com/lwch/logging"
	"github.com/lwch/runtime"
	"google.golang.org/protobuf/proto"
)

type desktopInfo struct {
	bits   int
	width  int
	height int
}

type Worker struct {
	workerOsBased
	info desktopInfo
}

func NewWorker() *Worker {
	worker := &Worker{}
	err := worker.init()
	if err != nil {
		return nil
	}
	return worker
}

func (worker *Worker) Do(conn *websocket.Conn) {
	defer conn.Close()
	for {
		_, data, err := conn.ReadMessage()
		runtime.Assert(err)
		var msg vncnetwork.VncMsg
		err = proto.Unmarshal(data, &msg)
		if err != nil {
			logging.Error("proto unmarshal: %v", err)
			continue
		}
		switch msg.GetXType() {
		case vncnetwork.VncMsg_capture_req:
			data := worker.runCapture()
			// if data.Ok {
			// 	dumpImage(data.Data, int(data.Width), int(data.Height))
			// }
			if !data.Ok {
				logging.Error("capture: %s", data.Msg)
			}
			msg.XType = vncnetwork.VncMsg_capture_data
			msg.Payload = &vncnetwork.VncMsg_Data{
				Data: &data,
			}
			enc, _ := proto.Marshal(&msg)
			conn.WriteMessage(websocket.BinaryMessage, enc)
		}
	}
}

func (worker *Worker) TestCapture() {
	msg := worker.runCapture()
	dumpImage(msg.Data, int(msg.Width), int(msg.Height))
}

func dumpImage(data []byte, width, height int) {
	f, err := os.Create(`C:\Users\lwch\Pictures\debug.jpeg`)
	if err != nil {
		logging.Error("debug: %v", err)
		return
	}
	defer f.Close()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	copy(img.Pix, data)
	err = jpeg.Encode(f, img, nil)
	if err != nil {
		logging.Error("encode: %v", err)
		return
	}
}
