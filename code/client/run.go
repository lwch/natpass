package main

import (
	"context"
	"natpass/code/client/global"
	"natpass/code/client/handler"
	"natpass/code/network"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
	"google.golang.org/grpc/metadata"
)

func makeContext(secret, id string) context.Context {
	return metadata.AppendToOutgoingContext(context.Background(),
		"secret", secret,
		"id", id)
}

func handshake(cfg *global.Configure, cli network.Natpass_ControlClient) error {
	return cli.Send(&network.ControlData{
		XType: network.ControlData_come,
		From:  cfg.ID,
	})
}

func run(cfg *global.Configure, cli network.NatpassClient) {
	ctx := makeContext(cfg.Secret, cfg.ID)

	ctlCli, err := cli.Control(ctx)
	runtime.Assert(err)
	fwdCli, err := cli.Forward(ctx)
	runtime.Assert(err)

	err = handshake(cfg, ctlCli)
	runtime.Assert(err)
	logging.Info("connect to server %s ok", cfg.Server)

	h := handler.New(ctx, ctlCli, fwdCli)

	h.CreateTunnels(cfg.ID, cfg.Tunnels)
	go h.HandleControl()

	h.HandleForward()
}
