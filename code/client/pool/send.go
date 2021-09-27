package pool

import (
	"natpass/code/client/global"
	"natpass/code/network"
	"time"
)

// SendConnectReq send connect request message
func (conn *Conn) SendConnectReq(id string, cfg global.Tunnel) {
	tp := network.ConnectRequest_tcp
	if cfg.Type != "tcp" {
		tp = network.ConnectRequest_udp
	}
	var msg network.Msg
	msg.To = cfg.Target
	msg.XType = network.Msg_connect_req
	msg.LinkId = id
	msg.Payload = &network.Msg_Creq{
		Creq: &network.ConnectRequest{
			Name:  cfg.Name,
			XType: tp,
			Addr:  cfg.RemoteAddr,
			Port:  uint32(cfg.RemotePort),
		},
	}
	select {
	case conn.write <- &msg:
	case <-time.After(conn.parent.cfg.WriteTimeout):
	}
}

// SendConnectError send connect error response message
func (conn *Conn) SendConnectError(to string, toIdx uint32, id, info string) {
	var msg network.Msg
	msg.To = to
	msg.ToIdx = toIdx
	msg.XType = network.Msg_connect_rep
	msg.LinkId = id
	msg.Payload = &network.Msg_Crep{
		Crep: &network.ConnectResponse{
			Ok:  false,
			Msg: info,
		},
	}
	select {
	case conn.write <- &msg:
	case <-time.After(conn.parent.cfg.WriteTimeout):
	}
}

// SendConnectOK send connect success response message
func (conn *Conn) SendConnectOK(to string, toIdx uint32, id string) {
	var msg network.Msg
	msg.To = to
	msg.ToIdx = toIdx
	msg.XType = network.Msg_connect_rep
	msg.LinkId = id
	msg.Payload = &network.Msg_Crep{
		Crep: &network.ConnectResponse{
			Ok: true,
		},
	}
	select {
	case conn.write <- &msg:
	case <-time.After(conn.parent.cfg.WriteTimeout):
	}
}

// SendDisconnect send disconnect message
func (conn *Conn) SendDisconnect(to string, toIdx uint32, id string) {
	var msg network.Msg
	msg.To = to
	msg.ToIdx = toIdx
	msg.XType = network.Msg_disconnect
	msg.LinkId = id
	select {
	case conn.write <- &msg:
	case <-time.After(conn.parent.cfg.WriteTimeout):
	}
}

// SendData send forward data
func (conn *Conn) SendData(to string, toIdx uint32, id string, data []byte) {
	dup := func(data []byte) []byte {
		ret := make([]byte, len(data))
		copy(ret, data)
		return ret
	}
	var msg network.Msg
	msg.To = to
	msg.ToIdx = toIdx
	msg.XType = network.Msg_forward
	msg.LinkId = id
	msg.Payload = &network.Msg_XData{
		XData: &network.Data{
			Data: dup(data),
		},
	}
	select {
	case conn.write <- &msg:
	case <-time.After(conn.parent.cfg.WriteTimeout):
	}
}

// SendKeepalive send keepalive message
func (conn *Conn) SendKeepalive() {
	var msg network.Msg
	msg.To = "server"
	msg.XType = network.Msg_keepalive
	select {
	case conn.write <- &msg:
	case <-time.After(conn.parent.cfg.WriteTimeout):
	}
}

// SendShellCreate send shell create message
func (conn *Conn) SendShellCreate(to, id, exec string) {
	var msg network.Msg
	msg.To = to
	msg.XType = network.Msg_shell_create
	msg.LinkId = id
	msg.Payload = &network.Msg_Screate{
		Screate: &network.ShellCreate{
			Exec: exec,
		},
	}
	select {
	case conn.write <- &msg:
	case <-time.After(conn.parent.cfg.WriteTimeout):
	}
}
