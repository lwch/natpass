package tunnel

import (
	"context"
	"natpass/code/client/global"
	"natpass/code/network"

	"github.com/lwch/runtime"
)

type Tunnel struct {
	ctx context.Context
	cli network.Natpass_ForwardClient

	// runtime
	local  string // local channel id
	remote string // remote channel id
}

func New(ctx context.Context, cfg global.Tunnel,
	ctlCli network.Natpass_ControlClient,
	fwdCli network.Natpass_ForwardClient) *Tunnel {
	cid, err := runtime.UUID(32)
	runtime.Assert(err)
	t := &Tunnel{
		ctx:   ctx,
		cli:   fwdCli,
		local: cid,
	}
	t.sendConnect(ctlCli, cfg)
	return t
}

func (t *Tunnel) sendConnect(cli network.Natpass_ControlClient, cfg global.Tunnel) {
	tpy := network.ConnectRequest_tcp
	if cfg.Type != "tcp" {
		tpy = network.ConnectRequest_udp
	}
	err := cli.Send(&network.ControlData{
		XType: network.ControlData_connect,
		Payload: &network.ControlData_Creq{
			Creq: &network.ConnectRequest{
				Name:  cfg.Name,
				Cid:   t.local,
				XType: tpy,
				Addr:  cfg.RemoteAddr,
				Port:  uint32(cfg.RemotePort),
			},
		},
	})
	runtime.Assert(err)
}

func (t *Tunnel) Run() {
	for {
		t.cli.Recv()
	}
}
