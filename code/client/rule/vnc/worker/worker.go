package worker

import (
	"image"
	"image/jpeg"
	"os"

	"github.com/gorilla/websocket"
	"github.com/lwch/logging"
	"github.com/lwch/natpass/code/client/rule/vnc/vncnetwork"
	"github.com/lwch/rdesktop"
	"github.com/lwch/runtime"
	"google.golang.org/protobuf/proto"
)

// Worker worker object
type Worker struct {
	cli *rdesktop.Client
}

// NewWorker create worker
func NewWorker(showCursor bool) *Worker {
	worker := &Worker{}
	cli, err := rdesktop.New()
	if err != nil {
		logging.Error("create rdesktop: %v", err)
		return nil
	}
	worker.cli = cli
	return worker
}

// Do handle worker
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
		case vncnetwork.VncMsg_mouse_event:
			worker.runMouse(msg.GetMouse())
		case vncnetwork.VncMsg_keyboard_event:
			worker.runKeyboard(msg.GetKeyboard())
		case vncnetwork.VncMsg_set_cursor:
			worker.cli.ShowCursor(msg.GetShowCursor())
		case vncnetwork.VncMsg_scroll_event:
			worker.runScroll(msg.GetScroll())
		case vncnetwork.VncMsg_clipboard_event:
			worker.runClipboard(conn, msg.GetClipboard())
		}
	}
}

// TestCapture test capture
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
