package client

import (
	"natpass/code/client/global"
	"natpass/code/network"
	"time"
)

func (c *Client) writeHandshake() error {
	var msg network.Msg
	msg.XType = network.Msg_handshake
	msg.From = c.cfg.ID
	msg.To = "server"
	msg.Payload = &network.Msg_Hsp{
		Hsp: &network.HandshakePayload{
			Enc: c.cfg.Enc[:],
		},
	}
	return c.conn.WriteMessage(&msg, 5*time.Second)
}

func (c *Client) sendConnect(id string, t global.Tunnel) {
	tp := network.ConnectRequest_tcp
	if t.Type != "tcp" {
		tp = network.ConnectRequest_udp
	}
	var msg network.Msg
	msg.From = c.cfg.ID
	msg.To = t.Target
	msg.XType = network.Msg_connect_req
	msg.Payload = &network.Msg_Creq{
		Creq: &network.ConnectRequest{
			Id:    id,
			Name:  t.Name,
			XType: tp,
			Addr:  t.RemoteAddr,
			Port:  uint32(t.RemotePort),
		},
	}
	c.conn.WriteMessage(&msg, 5*time.Second)
}

func (c *Client) connectError(to, id, m string) {
	var msg network.Msg
	msg.From = c.cfg.ID
	msg.To = to
	msg.XType = network.Msg_connect_rep
	msg.Payload = &network.Msg_Crep{
		Crep: &network.ConnectResponse{
			Id:  id,
			Ok:  false,
			Msg: m,
		},
	}
	c.conn.WriteMessage(&msg, time.Second)
}

func (c *Client) connectOK(to, id string) {
	var msg network.Msg
	msg.From = c.cfg.ID
	msg.To = to
	msg.XType = network.Msg_connect_rep
	msg.Payload = &network.Msg_Crep{
		Crep: &network.ConnectResponse{
			Id: id,
			Ok: true,
		},
	}
	c.conn.WriteMessage(&msg, time.Second)
}

func (c *Client) send(id, target string, data []byte) {
	var msg network.Msg
	msg.From = c.cfg.ID
	msg.To = target
	msg.XType = network.Msg_forward
	msg.Payload = &network.Msg_XData{
		XData: &network.Data{
			Cid:  id,
			Data: data,
		},
	}
	c.conn.WriteMessage(&msg, time.Second)
}

func (c *Client) disconnect(id, to string) {
	var msg network.Msg
	msg.From = c.cfg.ID
	msg.To = to
	msg.XType = network.Msg_disconnect
	msg.Payload = &network.Msg_XDisconnect{
		XDisconnect: &network.Disconnect{
			Id: id,
		},
	}
	c.conn.WriteMessage(&msg, time.Second)
	c.Lock()
	delete(c.links, id)
	c.Unlock()
}

func (c *Client) sendKeepalive() {
	var msg network.Msg
	msg.From = c.cfg.ID
	msg.To = "server"
	msg.XType = network.Msg_keepalive
	c.conn.WriteMessage(&msg, time.Second)
}
