package pool

import (
	"natpass/code/network"
	"time"

	"google.golang.org/protobuf/proto"
)

// SendData send forward data
func (conn *Conn) SendData(to string, toIdx uint32, id string, data []byte) uint64 {
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
		data, _ := proto.Marshal(&msg)
		return uint64(len(data))
	case <-time.After(conn.parent.cfg.WriteTimeout):
		return 0
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
