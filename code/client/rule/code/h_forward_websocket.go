package code

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/lwch/logging"
	"github.com/lwch/natpass/code/network"
)

var upgrader = websocket.Upgrader{
	EnableCompression: true,
}

func (code *Code) handleWebsocket(workspace *Workspace, w http.ResponseWriter, r *http.Request) {
	reqID, err := workspace.SendConnect(r)
	if err != nil {
		logging.Error("send_connect: %v", err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer workspace.closeMessage(reqID)
	resp := workspace.onResponse(reqID)
	if resp == nil {
		logging.Error("waiting for [%s] [%s] no response for websocket, request_id=%d",
			workspace.id, workspace.name, reqID)
		http.Error(w, "no response", http.StatusInternalServerError)
		return
	}

	if resp.GetXType() != network.Msg_code_connect_response {
		logging.Error("got invalid message type [%s] [%s]: %s",
			workspace.id, workspace.name, resp.GetXType().String())
		http.Error(w, "invalid message type", http.StatusServiceUnavailable)
		return
	}

	response := resp.GetCsconnRep()
	if !response.GetOk() {
		logging.Error("can not create websocket connection [%s] [%s]: %s",
			workspace.id, workspace.name, response.GetMsg())
		http.Error(w, response.GetMsg(), http.StatusBadGateway)
		return
	}

	local, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logging.Error("upgrade websocket connection [%s] [%s]: %v",
			workspace.id, workspace.name, err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer local.Close()

	var wg sync.WaitGroup

	wg.Add(2)
	go workspace.ws2remote(&wg, reqID, local)
	go workspace.remote2ws(&wg, reqID, local)
	wg.Wait()
}
