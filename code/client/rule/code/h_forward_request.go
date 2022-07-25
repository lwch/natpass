package code

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/lwch/logging"
	"github.com/lwch/natpass/code/network"
)

func (code *Code) handleRequest(workspace *Workspace, w http.ResponseWriter, r *http.Request) {
	reqID, err := workspace.SendRequest(r)
	if err != nil {
		logging.Error("send: %v", err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer workspace.closeMessage(reqID)
	resp := workspace.onResponse(reqID)
	if resp == nil {
		logging.Error("waiting for [%s] [%s] no response, request_id=%d",
			workspace.id, workspace.name, reqID)
		http.Error(w, "no response", http.StatusInternalServerError)
		return
	}

	if resp.GetXType() != network.Msg_code_response_hdr {
		logging.Error("got invalid message type [%s] [%s]: %s",
			workspace.id, workspace.name, resp.GetXType().String())
		http.Error(w, "invalid message type", http.StatusServiceUnavailable)
		return
	}

	hdr := resp.GetCsrepHdr()
	for key, values := range hdr.GetHeader() {
		for _, v := range values.GetValues() {
			w.Header().Add(key, v)
		}
	}

	w.WriteHeader(int(hdr.GetCode()))

	var idx uint32
	for {
		msg := workspace.onResponse(reqID)
		if msg == nil {
			logging.Error("no response")
			http.Error(w, "no response", http.StatusBadGateway)
			return
		}
		if msg.GetXType() != network.Msg_code_response_body {
			logging.Error("got invalid message type [%s] [%s]: %s",
				workspace.id, workspace.name, resp.GetXType().String())
			http.Error(w, "invalid message type", http.StatusServiceUnavailable)
			return
		}
		resp := msg.GetCsrepBody()
		if resp.GetIndex() != idx {
			logging.Error("loss data [%s] [%s]", workspace.id, workspace.name)
			http.Error(w, "loss data", http.StatusResetContent)
			return
		}
		if resp.GetMask()&1 == 0 {
			logging.Error("read data [%s] [%s]: %s", workspace.id, workspace.name, string(resp.GetBody()))
			http.Error(w, fmt.Sprintf("read error: %s", string(resp.GetBody())), http.StatusResetContent)
			return
		}
		_, err = io.Copy(w, bytes.NewReader(resp.GetBody()))
		if err != nil {
			logging.Error("write body: %v", err)
			return
		}
		if resp.GetMask()&2 > 0 {
			return
		}
		idx++
	}
}
