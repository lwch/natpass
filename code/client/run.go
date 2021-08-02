package main

import (
	"context"
	"natpass/code/client/global"
	"natpass/code/client/tunnel"
	"natpass/code/network"

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

	req := make(map[string]string, len(cfg.Tunnels))
	cfgs := make(map[string]global.Tunnel, len(cfg.Tunnels))
	for _, t := range cfg.Tunnels {
		id := createTunnel(ctx, t, ctlCli, cfg.ID)
		req[t.Name] = id
		cfgs[t.Name] = t
	}

	mgr := tunnel.NewMgr()
	go handleControl(ctlCli, req, cfgs, mgr)

	for {
		fwdCli.Recv()
	}
}

func handleControl(cli network.Natpass_ControlClient, req map[string]string,
	cfgs map[string]global.Tunnel, mgr *tunnel.Mgr) {
	for {
		data, err := cli.Recv()
		if err != nil {
			logging.Error("handler: %v", err)
			return
		}
		switch data.GetXType() {
		case network.ControlData_connect_req:
			req := data.GetCreq()
			cid, err := runtime.UUID(cidLength)
			if err != nil {
				logging.Error("generate channel_id failed, err=%v", err)
				connectFailed(cli, req.GetName(), err.Error(), data.GetTo(), data.GetFrom())
				continue
			}
			t := "tcp"
			if req.GetXType() == network.ConnectRequest_udp {
				t = "udp"
			}
			tn := tunnel.NewConnect(req.GetName(), cid, req.GetCid(), t, req.GetAddr(), req.GetPort())
			mgr.Add(tn)
			connectOK(cli, req.GetName(), cid, data.GetTo(), data.GetFrom())
		case network.ControlData_connect_rep:
		}
	}
}
