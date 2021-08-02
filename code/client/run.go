package main

import (
	"context"
	"natpass/code/client/global"
	"natpass/code/network"

	"github.com/lwch/runtime"
	"google.golang.org/grpc/metadata"
)

func makeContext(id, secret string) context.Context {
	return metadata.AppendToOutgoingContext(context.Background(),
		"id", id,
		"secret", secret)
}

func run(cfg *global.Configure, cli network.NatpassClient) {
	ctx := makeContext(cfg.ID, cfg.Secret)
	ctlCli, err := cli.Control(ctx)
	runtime.Assert(err)
	err = ctlCli.Send(&network.ControlData{
		XType: network.ControlData_come,
	})
	runtime.Assert(err)
	select {}
}
