package main

import (
	"context"
	"natpass/code/client/global"
	"natpass/code/client/tunnel"
	"natpass/code/network"
	"sync"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
	"google.golang.org/grpc/metadata"
)

func makeContext(id, secret string) context.Context {
	return metadata.AppendToOutgoingContext(context.Background(),
		"id", id,
		"secret", secret)
}

func handshake(cfg *global.Configure, cli network.Natpass_ControlClient) error {
	return cli.Send(&network.ControlData{
		XType: network.ControlData_come,
	})
}

func run(cfg *global.Configure, cli network.NatpassClient) {
	ctx := makeContext(cfg.ID, cfg.Secret)

	ctlCli, err := cli.Control(ctx)
	runtime.Assert(err)
	fwdCli, err := cli.Forward(ctx)
	runtime.Assert(err)

	err = handshake(cfg, ctlCli)
	runtime.Assert(err)
	logging.Info("connect to server %s ok", cfg.Server)

	var wg sync.WaitGroup
	wg.Add(len(cfg.Tunnels))
	for _, t := range cfg.Tunnels {
		tn := tunnel.New(ctx, t, ctlCli, fwdCli)
		go func(t global.Tunnel) {
			defer func() {
				wg.Done()
				logging.Error("tunnel for %s closed", t.Name)
			}()
			tn.Run()
		}(t)
	}

	wg.Wait()
}
