package code

import (
	"net/http"

	"github.com/lwch/logging"
)

func (code *Code) handleWebsocket(workspace *Workspace, w http.ResponseWriter, r *http.Request) {
	reqID, err := workspace.SendConnect(r)
	if err != nil {
		logging.Error("send_connect: %v", err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	logging.Info("reqID: %d", reqID)
}
