package main

import (
	"context"
	"natpass/code/client/global"
	"natpass/code/client/tunnel"
	"natpass/code/network"
	"net"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
	"google.golang.org/grpc/metadata"
)

func makeContext(secret string) context.Context {
	return metadata.AppendToOutgoingContext(context.Background(), "secret", secret)
}

func handshake(cfg *global.Configure, cli network.Natpass_ControlClient) error {
	return cli.Send(&network.ControlData{
		XType: network.ControlData_come,
		From:  cfg.ID,
	})
}

func run(cfg *global.Configure, cli network.NatpassClient) {
	ctx := makeContext(cfg.Secret)

	ctlCli, err := cli.Control(ctx)
	runtime.Assert(err)
	fwdCli, err := cli.Forward(ctx)
	runtime.Assert(err)

	err = handshake(cfg, ctlCli)
	runtime.Assert(err)
	logging.Info("connect to server %s ok", cfg.Server)

	cfgs := make(map[string]global.Tunnel, len(cfg.Tunnels)) // name => config
	conns := make(map[string]net.Conn)                       // local_cid => connection
	for _, t := range cfg.Tunnels {
		l, err := tunnel.NewListener(t.Type, t.Name, t.LocalAddr, t.LocalPort)
		runtime.Assert(err)
		cfgs[t.Name] = t
		go l.Loop(func(c net.Conn, name string) {
			id := createTunnel(ctx, cfgs[name], ctlCli, cfg.ID)
			conns[id] = c
		})
	}

	mgr := tunnel.NewMgr()
	go handleControl(ctlCli, conns, mgr)

	for {
		fwdCli.Recv()
	}
}

func handleControl(cli network.Natpass_ControlClient, conns map[string]net.Conn, mgr *tunnel.Mgr) {
	for {
		data, err := cli.Recv()
		if err != nil {
			logging.Error("handler: %v", err)
			return
		}
		switch data.GetXType() {
		case network.ControlData_connect_req:
			handleConnectReq(cli, data, mgr)
		case network.ControlData_connect_rep:
			handleConnectRep(cli, data, conns, mgr)
		}
	}
}

func handleConnectReq(cli network.Natpass_ControlClient, data *network.ControlData, mgr *tunnel.Mgr) {
	req := data.GetCreq()
	cid, err := runtime.UUID(cidLength)
	rep := &network.ConnectResponse{
		Name:      req.GetName(),
		RemoteCid: req.GetCid(),
	}
	if err != nil {
		logging.Error("generate channel_id failed, err=%v", err)
		rep.Ok = false
		rep.Msg = err.Error()
		connectResponse(cli, rep, data.GetTo(), data.GetFrom())
		return
	}
	t := "tcp"
	if req.GetXType() == network.ConnectRequest_udp {
		t = "udp"
	}
	tn, err := tunnel.NewConnect(req.GetName(), cid, req.GetCid(), t, req.GetAddr(), req.GetPort())
	if err != nil {
		logging.Error("new connect failed: %v", err)
		rep.Ok = false
		rep.Msg = err.Error()
		connectResponse(cli, rep, data.GetTo(), data.GetFrom())
		return
	}
	mgr.Add(tn)
	rep.Ok = true
	rep.LocalCid = cid
	connectResponse(cli, rep, data.GetTo(), data.GetFrom())
}

func handleConnectRep(cli network.Natpass_ControlClient, data *network.ControlData,
	conns map[string]net.Conn, mgr *tunnel.Mgr) {
	rep := data.GetCrep()
	if !rep.GetOk() {
		logging.Error("connect %s failed, err=%s", rep.GetName(), rep.GetMsg())
		return
	}
	tn, err := tunnel.NewListen(rep.GetName(), rep.GetRemoteCid(), rep.GetLocalCid(), conns[rep.GetRemoteCid()])
	runtime.Assert(err)
	mgr.Add(tn)
}
