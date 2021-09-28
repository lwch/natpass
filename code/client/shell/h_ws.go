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
	defer link.Close()
	<-link.onWork
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
	defer link.Close()
	for {
		msg := <-ch
		if msg == nil {
			return
		}
		link.SetTargetIdx(msg.GetFromIdx())
		switch msg.GetXType() {
		case network.Msg_shell_created:
			if msg.GetScreated().GetOk() {
				link.onWork <- struct{}{}
				continue
			}
			logging.Error("create shell %s on tunnel %s failed, err=%s",
				link.id, link.parent.Name, msg.GetScreated().GetMsg())
			return
		case network.Msg_shell_data:
			err := local.WriteMessage(websocket.TextMessage, msg.GetSdata().GetData())
			if err != nil {
				logging.Error("write data for %s failed: %v", shell.Name, err)
				return
			}
			logging.Debug("remote read %d bytes: name=%s, id=%s",
				len(msg.GetSdata().GetData()), shell.Name, id)
		case network.Msg_shell_close:
			logging.Info("shell %s on tunnel %s closed by remote",
				link.id, link.parent.Name)
			return
		}
	}
}
