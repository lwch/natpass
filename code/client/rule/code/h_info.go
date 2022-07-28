package code

import (
	"encoding/json"
	"net/http"

	"github.com/lwch/logging"
)

// Info get workspace info
func (code *Code) Info(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	code.RLock()
	workspace := code.workspace[id]
	code.RUnlock()
	if workspace == nil {
		http.NotFound(w, r)
		return
	}
	data, err := json.Marshal(map[string]interface{}{
		"name":        code.Name,
		"send_bytes":  workspace.sendBytes,
		"send_packet": workspace.sendPacket,
		"recv_bytes":  workspace.recvBytes,
		"recv_packet": workspace.recvPacket,
	})
	if err != nil {
		logging.Error("marshal: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
