package pool

import (
	"fmt"
	"natpass/code/client/global"
	"natpass/code/client/tunnel"
	"natpass/code/network"
	"net"
	"strconv"
	"time"

	"github.com/lwch/logging"
)

func (p *Pool) handleConnectReq(conn *network.Conn, from, to string, req *network.ConnectRequest) {
	dial := "tcp"
	if req.GetXType() == network.ConnectRequest_udp {
		dial = "udp"
	}
	link, err := net.Dial(dial, fmt.Sprintf("%s:%d", req.GetAddr(), req.GetPort()))
	if err != nil {
		p.sendConnectError(conn, to, from, req.GetId(), err.Error())
		return
	}
	host, pt, _ := net.SplitHostPort(link.LocalAddr().String())
	port, _ := strconv.ParseUint(pt, 10, 16)
	tn := tunnel.New(global.Tunnel{
		Name:       req.GetName(),
		Target:     from,
		Type:       dial,
		LocalAddr:  host,
		LocalPort:  uint16(port),
		RemoteAddr: req.GetAddr(),
		RemotePort: uint16(req.GetPort()),
	}, p)
	tn.NewLink(req.GetId(), from, link)
	p.Add(tn)
	p.sendConnectOK(conn, to, from, req.GetId())
}

func (p *Pool) handleConnectRep(rep *network.ConnectResponse) {
	id := rep.GetId()

	p.RLock()
	link := p.links[id]
	p.RUnlock()

	if link == nil {
		logging.Error("link %s not found", id)
		return
	}

	if rep.GetOk() {
		link.OnWork <- struct{}{}
	} else {
		link.Close()
	}
}

func (p *Pool) handleDisconnect(data *network.Disconnect) {
	id := data.GetId()

	p.RLock()
	link := p.links[id]
	p.RUnlock()

	if link != nil {
		logging.Info("disconnect link %s from remote", id)
		link.Close()
	}
}

func (p *Pool) handleData(data *network.Data) {
	id := data.GetLid()
	p.RLock()
	link := p.links[id]
	p.RUnlock()
	if link == nil {
		logging.Error("link %s not found", id)
		return
	}
	link.WriteData(data.GetData())
}

func (p *Pool) sendConnectError(conn *network.Conn, from, to, id, m string) {
	var msg network.Msg
	msg.From = from
	msg.To = to
	msg.XType = network.Msg_connect_rep
	msg.Payload = &network.Msg_Crep{
		Crep: &network.ConnectResponse{
			Id:  id,
			Ok:  false,
			Msg: m,
		},
	}
	conn.WriteMessage(&msg, time.Second)
}

func (c *Pool) sendConnectOK(conn *network.Conn, from, to, id string) {
	var msg network.Msg
	msg.From = from
	msg.To = to
	msg.XType = network.Msg_connect_rep
	msg.Payload = &network.Msg_Crep{
		Crep: &network.ConnectResponse{
			Id: id,
			Ok: true,
		},
	}
	conn.WriteMessage(&msg, time.Second)
}
