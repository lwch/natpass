package worker

import (
	"natpass/code/client/tunnel/vnc/network"

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
		var msg network.VncMsg
		err = proto.Unmarshal(data, &msg)
		if err != nil {
			logging.Error("proto unmarshal: %v", err)
			continue
		}
		switch msg.GetXType() {
		case network.VncMsg_capture_req:
			worker.runCapture(conn)
		}
	}
}
