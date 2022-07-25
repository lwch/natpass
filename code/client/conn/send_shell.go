package conn

import (
	"time"

	"github.com/lwch/natpass/code/network"
	"google.golang.org/protobuf/proto"
)

// SendShellData send shell data
func (conn *Conn) SendShellData(to string, id string, data []byte) uint64 {
	var msg network.Msg
	msg.To = to
	msg.XType = network.Msg_shell_data
	msg.LinkId = id
	msg.Payload = &network.Msg_Sdata{
		Sdata: &network.ShellData{
			Data: dup(data),
		},
	}
	select {
	case conn.write <- &msg:
		data, _ := proto.Marshal(&msg)
		return uint64(len(data))
	case <-time.After(conn.cfg.WriteTimeout):
		return 0
	}
}

// SendShellResize send shell resize
func (conn *Conn) SendShellResize(to string, id string, rows, cols uint32) {
	var msg network.Msg
	msg.To = to
	msg.XType = network.Msg_shell_resize
	msg.LinkId = id
	msg.Payload = &network.Msg_Sresize{
		Sresize: &network.ShellResize{
			Rows: rows,
			Cols: cols,
		},
	}
	select {
	case conn.write <- &msg:
	case <-time.After(conn.cfg.WriteTimeout):
	}
}
