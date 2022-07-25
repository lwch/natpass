package code

import (
	"bytes"
	"io"
	"net/http"

	"github.com/lwch/logging"
	"github.com/lwch/natpass/code/network"
	"github.com/lwch/natpass/code/utils"
)

func (ws *Workspace) handleRequest(msg *network.Msg) {
	defer utils.Recover("handleRequest")
	req := msg.GetCsreq()
	request, err := http.NewRequest(req.GetMethod(), "http://unix"+req.GetUri(), bytes.NewReader(req.GetBody()))
	if err != nil {
		logging.Error("build request [%s] [%s]: %v", ws.id, ws.name, err)
		return
	}
	for key, values := range req.GetHeader() {
		for _, v := range values.GetValues() {
			request.Header.Add(key, v)
		}
	}
	response, err := ws.cli.Do(request)
	if err != nil {
		// TODO: response error
		logging.Error("call request [%s] [%s] [%s]: %v", ws.id, ws.name, req.GetUri(), err)
		return
	}
	defer response.Body.Close()
	send := ws.remote.SendCodeResponseHeader(ws.target, ws.id, req.GetRequestId(), uint32(response.StatusCode), response.Header)
	ws.sendBytes += send
	ws.sendPacket++
	buf := make([]byte, 32*1024) // 32k block read
	var idx uint32
	for {
		n, err := response.Body.Read(buf)
		if err != nil {
			if err == io.EOF {
				send := ws.remote.SendCodeResponseBody(ws.target, ws.id, req.GetRequestId(), idx, true, true, buf[:n])
				ws.sendBytes += send
				ws.sendPacket++
				idx++
				return
			}
			send := ws.remote.SendCodeResponseBody(ws.target, ws.id, req.GetRequestId(), idx, false, true, []byte(err.Error()))
			ws.sendBytes += send
			ws.sendPacket++
			idx++
			logging.Error("call request [%s] [%s] [%s] read response data: %v", ws.id, ws.name, req.GetUri(), err)
			return
		}
		send := ws.remote.SendCodeResponseBody(ws.target, ws.id, req.GetRequestId(), idx, true, false, buf[:n])
		ws.sendBytes += send
		ws.sendPacket++
		idx++
	}
}
