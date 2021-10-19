package main

import (
	"fmt"
	"natpass/code/client/global"
	"natpass/code/client/pool"
	"natpass/code/client/tunnel"
	"natpass/code/client/tunnel/reverse"
	"natpass/code/client/tunnel/shell"
	"natpass/code/client/tunnel/vnc"
	"natpass/code/network"
	"net"
	"strconv"

	"github.com/lwch/logging"
)

func connect(mgr *tunnel.Mgr, conn *pool.Conn, msg *network.Msg) {
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
	tn := mgr.Get(req.GetName(), msg.GetFrom())
	if tn == nil {
		tn = reverse.New(global.Tunnel{
			Name:       req.GetName(),
			Target:     msg.GetFrom(),
			Type:       dial,
			LocalAddr:  host,
			LocalPort:  uint16(port),
			RemoteAddr: addr.GetAddr(),
			RemotePort: uint16(addr.GetPort()),
		})
		mgr.Add(tn)
	}
	lk := tn.NewLink(msg.GetLinkId(), msg.GetFrom(), msg.GetFromIdx(), link, conn).(*reverse.Link)
	conn.SendConnectOK(msg.GetFrom(), msg.GetFromIdx(), msg.GetLinkId())
	lk.Forward()
	lk.OnWork <- struct{}{}
}

func shellCreate(mgr *tunnel.Mgr, conn *pool.Conn, msg *network.Msg) {
	create := msg.GetCreq()
	tn := mgr.Get(create.GetName(), msg.GetFrom())
	if tn == nil {
		tn = shell.New(global.Tunnel{
			Name:   create.GetName(),
			Target: msg.GetFrom(),
			Type:   "shell",
			Exec:   create.GetCshell().GetExec(),
			Env:    create.GetCshell().GetEnv(),
		})
		mgr.Add(tn)
	}
	lk := tn.NewLink(msg.GetLinkId(), msg.GetFrom(), msg.GetFromIdx(), nil, conn).(*shell.Link)
	err := lk.Exec()
	if err != nil {
		logging.Error("create shell failed: %v", err)
		conn.SendConnectError(msg.GetFrom(), msg.GetFromIdx(), msg.GetLinkId(), err.Error())
		return
	}
	conn.SendConnectOK(msg.GetFrom(), msg.GetFromIdx(), msg.GetLinkId())
	lk.Forward()
}

func vncCreate(mgr *tunnel.Mgr, conn *pool.Conn, msg *network.Msg) {
	create := msg.GetCreq()
	tn := mgr.Get(create.GetName(), msg.GetFrom())
	if tn == nil {
		tn = vnc.New(global.Tunnel{
			Name:   create.GetName(),
			Target: msg.GetFrom(),
			Type:   "vnc",
			Fps:    create.GetCvnc().GetFps(),
		})
		mgr.Add(tn)
	}
	lk := vnc.NewLink(tn.(*vnc.VNC), msg.GetLinkId(), msg.GetFrom(), conn)
	lk.SetTargetIdx(msg.GetFromIdx())
	err := lk.Fork()
	if err != nil {
		logging.Error("create vnc failed: %v", err)
		conn.SendConnectError(msg.GetFrom(), msg.GetFromIdx(), msg.GetLinkId(), err.Error())
		return
	}
	conn.SendConnectOK(msg.GetFrom(), msg.GetFromIdx(), msg.GetLinkId())
	lk.Forward()
}
