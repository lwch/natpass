package main

import (
	"fmt"
	"natpass/code/client/global"
	"natpass/code/client/pool"
	"natpass/code/client/shell"
	"natpass/code/client/tunnel"
	"natpass/code/network"
	"net"
	"strconv"

	"github.com/lwch/logging"
)

func connect(conn *pool.Conn, msg *network.Msg) {
	req := msg.GetCreq()
	dial := "tcp"
	if req.GetXType() == network.ConnectRequest_udp {
		dial = "udp"
	}
	addr := req.GetCaddr()
	link, err := net.Dial(dial, fmt.Sprintf("%s:%d", addr.GetAddr(), addr.GetPort()))
	if err != nil {
		logging.Error("connect to %s:%d failed, err=%v", addr.GetAddr(), addr.GetPort(), err)
		conn.SendConnectError(msg.GetFrom(), msg.GetFromIdx(), msg.GetLinkId(), err.Error())
		return
	}
	host, pt, _ := net.SplitHostPort(link.LocalAddr().String())
	port, _ := strconv.ParseUint(pt, 10, 16)
	tn := tunnel.New(global.Tunnel{
		Name:       req.GetName(),
		Target:     msg.GetFrom(),
		Type:       dial,
		LocalAddr:  host,
		LocalPort:  uint16(port),
		RemoteAddr: addr.GetAddr(),
		RemotePort: uint16(addr.GetPort()),
	})
	lk := tunnel.NewLink(tn, msg.GetLinkId(), msg.GetFrom(), link, conn)
	lk.SetTargetIdx(msg.GetFromIdx())
	conn.SendConnectOK(msg.GetFrom(), msg.GetFromIdx(), msg.GetLinkId())
	lk.Forward()
	lk.OnWork <- struct{}{}
}

func shellCreate(conn *pool.Conn, msg *network.Msg) {
	create := msg.GetCreq()
	sh := shell.New(global.Tunnel{
		Name:   create.GetName(),
		Target: msg.GetFrom(),
		Type:   "shell",
		Exec:   create.GetCshell().GetExec(),
		Env:    create.GetCshell().GetEnv(),
	})
	lk := shell.NewLink(sh, msg.GetLinkId(), msg.GetFrom(), conn)
	lk.SetTargetIdx(msg.GetFromIdx())
	err := lk.Exec()
	if err != nil {
		logging.Error("create shell failed: %v", err)
		conn.SendConnectError(msg.GetFrom(), msg.GetFromIdx(), msg.GetLinkId(), err.Error())
		return
	}
	conn.SendConnectOK(msg.GetFrom(), msg.GetFromIdx(), msg.GetLinkId())
	lk.Forward()
}
