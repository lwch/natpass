package code

import (
	"bytes"
	"io"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
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
		ws.remote.SendCodeResponseHeader(ws.target, ws.id, req.GetRequestId(),
			http.StatusInternalServerError, nil)
		ws.remote.SendCodeResponseBody(ws.target, ws.id, req.GetRequestId(),
			0, false, true, []byte(err.Error()))
		return
	}
	for key, values := range req.GetHeader() {
		for _, v := range values.GetValues() {
			request.Header.Add(key, v)
		}
	}
	response, err := ws.cli.Do(request)
	if err != nil {
		logging.Error("call request [%s] [%s] [%s]: %v", ws.id, ws.name, req.GetUri(), err)
		ws.remote.SendCodeResponseHeader(ws.target, ws.id, req.GetRequestId(),
			http.StatusInternalServerError, nil)
		ws.remote.SendCodeResponseBody(ws.target, ws.id, req.GetRequestId(),
			0, false, true, []byte(err.Error()))
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
				return
			}
			send := ws.remote.SendCodeResponseBody(ws.target, ws.id, req.GetRequestId(), idx, false, true, []byte(err.Error()))
			ws.sendBytes += send
			ws.sendPacket++
			logging.Error("call request [%s] [%s] [%s] read response data: %v", ws.id, ws.name, req.GetUri(), err)
			return
		}
		send := ws.remote.SendCodeResponseBody(ws.target, ws.id, req.GetRequestId(), idx, true, false, buf[:n])
		ws.sendBytes += send
		ws.sendPacket++
		idx++
	}
}

func (ws *Workspace) handleConnect(msg *network.Msg) {
	defer utils.Recover("handleConnect")

	connect := msg.GetCsconn()
	hdr := make(http.Header)
	for key, values := range connect.GetHeader() {
		for _, value := range values.GetValues() {
			hdr.Add(key, value)
		}
	}

	remote, resp, err := ws.dailer.Dial("ws://unix"+connect.GetUri(), hdr)
	if err != nil {
		logging.Error("dial websocket [%s] [%s]: %v", ws.id, ws.name, err)
		send := ws.remote.SendCodeResponseConnect(ws.target, ws.id, connect.GetRequestId(),
			false, err.Error(), nil)
		ws.sendBytes += send
		ws.sendPacket++
		return
	}
	defer remote.Close()
	defer resp.Body.Close()
	send := ws.remote.SendCodeResponseConnect(ws.target, ws.id, connect.GetRequestId(),
		true, "", resp.Header)
	ws.sendBytes += send
	ws.sendPacket++

	var wg sync.WaitGroup

	wg.Add(2)
	go ws.ws2remote(&wg, connect.GetRequestId(), remote)
	go ws.remote2ws(&wg, connect.GetRequestId(), remote)
	wg.Wait()
}

func (ws *Workspace) ws2remote(wg *sync.WaitGroup, reqID uint64, conn *websocket.Conn) {
	defer wg.Done()
	defer conn.Close()
	defer ws.closeMessage(reqID)
	for {
		t, data, err := conn.ReadMessage()
		if err != nil {
			logging.Error("read_message [%s] [%s]: %v", ws.id, ws.name, err)
			ws.SendData(reqID, false, websocket.TextMessage, []byte(err.Error()))
			return
		}
		ws.SendData(reqID, true, t, data)
	}
}

// SendData send websocket data
func (ws *Workspace) SendData(reqID uint64, ok bool, t int, body []byte) {
	for i := 0; i < len(body); i += 32 * 1024 {
		end := i + 32*1024
		if end > len(body) {
			end = len(body)
		}
		send := ws.remote.SendCodeData(ws.target, ws.id, reqID,
			ok, t, body[i:end])
		ws.sendBytes += send
		ws.sendPacket++
	}
}

func (ws *Workspace) remote2ws(wg *sync.WaitGroup, reqID uint64, conn *websocket.Conn) {
	defer wg.Done()
	defer conn.Close()
	defer ws.closeMessage(reqID)
	ch := ws.chanResponse(reqID)
	for {
		msg := <-ch
		if msg == nil {
			return
		}
		if msg.GetXType() != network.Msg_code_data {
			logging.Error("got invalid message type [%s] [%s]: %s",
				ws.id, ws.name, msg.GetXType().String())
			return
		}
		data := msg.GetCsdata()
		if !data.GetOk() {
			logging.Error("read message [%s] [%s]: %s",
				ws.id, ws.name, string(data.GetData()))
			return
		}
		err := conn.WriteMessage(int(data.GetType()), data.GetData())
		if err != nil {
			logging.Error("write_message [%s] [%s]: %v", ws.id, ws.name, err)
			return
		}
	}
}
