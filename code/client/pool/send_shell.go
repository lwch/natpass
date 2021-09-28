package pool

import (
	"natpass/code/client/global"
	"natpass/code/network"
	"time"
)

// SendShellCreate send shell create message
func (conn *Conn) SendShellCreate(id string, cfg global.Tunnel) {
	var msg network.Msg
	msg.To = cfg.Target
	msg.XType = network.Msg_shell_create
	msg.LinkId = id
	msg.Payload = &network.Msg_Screate{
		Screate: &network.ShellCreate{
			Name: cfg.Name,
			Exec: cfg.Exec,
		},
	}
	select {
	case conn.write <- &msg:
	case <-time.After(conn.parent.cfg.WriteTimeout):
	}
}

// SendShellCreatedError send shell create error response message
func (conn *Conn) SendShellCreatedError(to string, toIdx uint32, id, info string) {
	var msg network.Msg
	msg.To = to
	msg.ToIdx = toIdx
	msg.XType = network.Msg_shell_created
	msg.LinkId = id
	msg.Payload = &network.Msg_Screated{
		Screated: &network.ShellCreated{
			Ok:  false,
			Msg: info,
		},
	}
	select {
	case conn.write <- &msg:
	case <-time.After(conn.parent.cfg.WriteTimeout):
	}
}

// SendShellCreatedOK send shell create success response message
func (conn *Conn) SendShellCreatedOK(to string, toIdx uint32, id string) {
	var msg network.Msg
	msg.To = to
	msg.ToIdx = toIdx
	msg.XType = network.Msg_shell_created
	msg.LinkId = id
	msg.Payload = &network.Msg_Screated{
		Screated: &network.ShellCreated{
			Ok: true,
		},
	}
	select {
	case conn.write <- &msg:
	case <-time.After(conn.parent.cfg.WriteTimeout):
	}
}

// SendData send shell data
func (conn *Conn) SendShellData(to string, toIdx uint32, id string, data []byte) {
	dup := func(data []byte) []byte {
		ret := make([]byte, len(data))
		copy(ret, data)
		return ret
	}
	var msg network.Msg
	msg.To = to
	msg.ToIdx = toIdx
	msg.XType = network.Msg_shell_data
	msg.LinkId = id
	msg.Payload = &network.Msg_Sdata{
		Sdata: &network.ShellData{
			Data: dup(data),
		},
	}
	select {
	case conn.write <- &msg:
	case <-time.After(conn.parent.cfg.WriteTimeout):
	}
}
