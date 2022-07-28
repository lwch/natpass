package shell

import (
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/lwch/logging"
	"github.com/lwch/natpass/code/client/conn"
	"github.com/lwch/natpass/code/network"
	"github.com/lwch/natpass/code/utils"
	"google.golang.org/protobuf/proto"
)

var upgrader = websocket.Upgrader{}

// WS websocket for forward data
func (shell *Shell) WS(conn *conn.Conn, w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/ws/")

	local, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logging.Error("upgrade websocket failed: %s, err=%v", shell.Name, err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer local.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		shell.localForward(id, local)
	}()
	go func() {
		defer wg.Done()
		shell.remoteForward(id, local)
	}()
	wg.Wait()
}

func (shell *Shell) localForward(id string, local *websocket.Conn) {
	defer utils.Recover("localForward")
	defer local.Close()
	shell.RLock()
	link := shell.links[id]
	shell.RUnlock()
	defer link.Close(true)
	for {
		_, data, err := local.ReadMessage()
		if err != nil {
			logging.Error("read local data for %s failed: %v", shell.Name, err)
			return
		}
		link.SendData(data)
		logging.Debug("local read %d bytes: name=%s, id=%s", len(data), shell.Name, id)
	}
}

func (shell *Shell) remoteForward(id string, local *websocket.Conn) {
	defer utils.Recover("remoteForward")
	defer local.Close()
	shell.RLock()
	link := shell.links[id]
	shell.RUnlock()
	ch := link.remote.ChanRead(id)
	defer link.Close(true)
	for {
		msg := <-ch
		if msg == nil {
			return
		}
		data, _ := proto.Marshal(msg)
		link.recvBytes += uint64(len(data))
		link.recvPacket++
		switch msg.GetXType() {
		case network.Msg_shell_data:
			err := local.WriteMessage(websocket.TextMessage, msg.GetSdata().GetData())
			if err != nil {
				logging.Error("write data for %s failed: %v", shell.Name, err)
				return
			}
			logging.Debug("remote read %d bytes: name=%s, id=%s",
				len(msg.GetSdata().GetData()), shell.Name, id)
		}
	}
}
