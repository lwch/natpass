package code

import (
	"io/ioutil"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/lwch/logging"
	"github.com/lwch/natpass/code/network"
)

// SendRequest send request from local node
func (ws *Workspace) SendRequest(r *http.Request) (uint64, error) {
	reqID := atomic.AddUint64(&ws.requestID, 1)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logging.Error("send request for workspace [%s] [%s]: %v", ws.id, ws.name, err)
		return 0, err
	}
	ws.Lock()
	ws.onMessage[reqID] = make(chan *network.Msg, 1024)
	ws.Unlock()
	send := ws.remote.SendCodeRequest(ws.target, ws.id, reqID,
		r.Method, r.URL.RequestURI(), body, r.Header)
	ws.sendBytes += send
	ws.sendPacket++
	return reqID, nil
}

// SendConnect send websocket connect action from local node
func (ws *Workspace) SendConnect(r *http.Request) (uint64, error) {
	reqID := atomic.AddUint64(&ws.requestID, 1)
	ws.Lock()
	ws.onMessage[reqID] = make(chan *network.Msg, 1024)
	ws.Unlock()

	hdr := make(http.Header)
	for key, values := range r.Header {
		if strings.HasPrefix(key, "Sec-") {
			continue
		}
		for _, value := range values {
			hdr.Add(key, value)
		}
	}

	hdr.Del("Connection")
	hdr.Del("Upgrade")

	send := ws.remote.SendCodeConnect(ws.target, ws.id, reqID,
		r.URL.RequestURI(), hdr)
	ws.sendBytes += send
	ws.sendPacket++
	return reqID, nil
}
