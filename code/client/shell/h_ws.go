package shell

import (
	"natpass/code/client/pool"
	"natpass/code/network"
	"natpass/code/utils"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/lwch/logging"
)

var upgrader = websocket.Upgrader{}

// WS websocket for forward data
func (shell *Shell) WS(pool *pool.Pool, w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/ws/")

	conn := pool.Get(id)
	logging.Info("ws: %d", conn.Idx)
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
		shell.localForward(id, local, conn)
	}()
	go func() {
		defer wg.Done()
		shell.remoteForward(id, conn.ChanRead(id), local)
	}()
	wg.Wait()
}

func (shell *Shell) localForward(id string, local *websocket.Conn, remote *pool.Conn) {
	defer utils.Recover("localForward")
	defer local.Close()
	logging.Info("localForward: %d", remote.Idx)
	for {
		_, data, err := local.ReadMessage()
		if err != nil {
			// TODO: close
			logging.Error("read local data for %s failed: %v", shell.Name, err)
			return
		}
		remote.SendShellData(shell.cfg.Target, remote.Idx, id, data)
		logging.Info("send: %d", remote.Idx)
		logging.Debug("local read %d bytes: name=%s, id=%s", len(data), shell.Name, id)
	}
}

func (shell *Shell) remoteForward(id string, ch <-chan *network.Msg, local *websocket.Conn) {
	defer utils.Recover("remoteForward")
	defer local.Close()
	for {
		msg := <-ch
		if msg == nil {
			return
		}
		switch msg.GetXType() {
		case network.Msg_shell_data:
			err := local.WriteMessage(websocket.TextMessage, msg.GetSdata().GetData())
			if err != nil {
				logging.Error("write data for %s failed: %v", shell.Name, err)
				return
			}
			logging.Debug("remote read %d bytes: name=%s, id=%s",
				len(msg.GetSdata().GetData()), shell.Name, id)
			// TODO: other
		}
	}
}
