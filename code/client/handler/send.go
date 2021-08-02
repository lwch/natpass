package handler

import (
	"context"
	"natpass/code/client/global"
	"natpass/code/network"

	"github.com/lwch/runtime"
)

const cidLength = 32

func createTunnel(ctx context.Context, cfg global.Tunnel, cli network.Natpass_ControlClient, id string) string {
	cid, err := runtime.UUID(cidLength)
	runtime.Assert(err)
	sendConnect(cli, cid, cfg, id, cfg.Target)
	return cid
}

func sendConnect(cli network.Natpass_ControlClient, cid string, cfg global.Tunnel,
	from, to string) {
	tpy := network.ConnectRequest_tcp
	if cfg.Type != "tcp" {
		tpy = network.ConnectRequest_udp
	}
	err := cli.Send(&network.ControlData{
		XType: network.ControlData_connect_req,
		From:  from,
		To:    to,
		Payload: &network.ControlData_Creq{
			Creq: &network.ConnectRequest{
				Name:  cfg.Name,
				Cid:   cid,
				XType: tpy,
				Addr:  cfg.RemoteAddr,
				Port:  uint32(cfg.RemotePort),
			},
		},
	})
	runtime.Assert(err)
}

func connectResponse(cli network.Natpass_ControlClient, payload *network.ConnectResponse, from, to string) {
	cli.Send(&network.ControlData{
		XType: network.ControlData_connect_rep,
		From:  from,
		To:    to,
		Payload: &network.ControlData_Crep{
			Crep: payload,
		},
	})
}
