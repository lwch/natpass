package tunnel

import (
	"natpass/code/client/global"
	"natpass/code/network"
)

func (link *Link) sendConnect(id string, t global.Tunnel) {
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
	link.write <- &msg
}

func (link *Link) sendDisconnect(id, to string) {
	var msg network.Msg
	msg.To = to
	msg.XType = network.Msg_disconnect
	msg.Payload = &network.Msg_XDisconnect{
		XDisconnect: &network.Disconnect{
			Id: id,
		},
	}
	link.write <- &msg
}

func (link *Link) sendData(id, target string, data []byte) {
	var msg network.Msg
	msg.To = target
	msg.XType = network.Msg_forward
	msg.Payload = &network.Msg_XData{
		XData: &network.Data{
			Lid:  id,
			Data: data,
		},
	}
	link.write <- &msg
}
