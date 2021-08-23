package pool

import (
	"natpass/code/client/global"
	"natpass/code/network"
)

// SendConnect send connect request message
func (p *Pool) SendConnect(id string, t global.Tunnel) {
	tp := network.ConnectRequest_tcp
	if t.Type != "tcp" {
		tp = network.ConnectRequest_udp
	}
	var msg network.Msg
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
	p.write <- &msg
}

// SendDisconnect send disconnect message
func (p *Pool) SendDisconnect(id, to string) {
	var msg network.Msg
	msg.To = to
	msg.XType = network.Msg_disconnect
	msg.Payload = &network.Msg_XDisconnect{
		XDisconnect: &network.Disconnect{
			Id: id,
		},
	}
	p.write <- &msg
}

// SendData send forward data
func (p *Pool) SendData(id, to string, data []byte) {
	dup := func(data []byte) []byte {
		ret := make([]byte, len(data))
		copy(ret, data)
		return ret
	}
	var msg network.Msg
	msg.To = to
	msg.XType = network.Msg_forward
	msg.Payload = &network.Msg_XData{
		XData: &network.Data{
			Lid:  id,
			Data: dup(data),
		},
	}
	p.write <- &msg
}
