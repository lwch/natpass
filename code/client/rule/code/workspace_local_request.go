package code

import (
	"io/ioutil"
	"net/http"
	"sync/atomic"

	"github.com/lwch/logging"
	"github.com/lwch/natpass/code/network"
)

func (ws *Workspace) SendRequest(r *http.Request) (uint64, error) {
	reqID := atomic.AddUint64(&ws.requestID, 1)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logging.Error("send request for workspace [%s] [%s]: %v", ws.id, ws.name, err)
		return 0, err
	}
	ws.Lock()
	ws.onMessage[reqID] = make(chan *network.Msg)
	ws.Unlock()
	send := ws.remote.SendCodeRequest(ws.target, ws.id, reqID,
		r.Method, r.URL.RequestURI(), body, r.Header)
	ws.sendBytes += send
	ws.sendPacket++
	return reqID, nil
}
