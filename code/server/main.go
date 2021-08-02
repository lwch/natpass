package main

import (
	"flag"
	"fmt"
	"natpass/code/network"
	"natpass/code/server/global"
	"natpass/code/server/server"
	"net"
	"os"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	conf := flag.String("conf", "", "configure file path")
	flag.Parse()

	if len(*conf) == 0 {
		fmt.Println("missing -conf param")
		os.Exit(1)
	}

	cfg := global.LoadConf(*conf)

	tls, err := credentials.NewServerTLSFromFile(cfg.TLSCrt, cfg.TLSKey)
	runtime.Assert(err)

	handler := server.NewHandler(cfg)

	svr := grpc.NewServer(grpc.Creds(tls))
	network.RegisterNatpassServer(svr, handler)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Listen))
	runtime.Assert(err)

	logging.Info("listen on %d", cfg.Listen)

	runtime.Assert(svr.Serve(listener))
}
