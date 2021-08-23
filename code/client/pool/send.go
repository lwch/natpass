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
	p.writeConnect <- connectData{
		to:   t.Target,
		id:   id,
		name: t.Name,
		tp:   tp,
		addr: t.RemoteAddr,
		port: t.RemotePort,
	}
}

// SendDisconnect send disconnect message
func (p *Pool) SendDisconnect(id, to string) {
	p.writeDisconnect <- disconnectData{
		to: to,
		id: id,
	}
}

// SendData send forward data
func (p *Pool) SendData(id, to string, data []byte) {
	dup := func(data []byte) []byte {
		ret := make([]byte, len(data))
		copy(ret, data)
		return ret
	}
	p.writeForward <- forwardData{
		to:   to,
		id:   id,
		data: dup(data),
	}
}
