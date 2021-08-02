package main

import (
	"flag"
	"fmt"
	"natpass/code/client/global"
	"natpass/code/network"
	"os"

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

	conn, err := grpc.Dial(cfg.Server,
		grpc.WithTransportCredentials(credentials.NewTLS(nil)))
	runtime.Assert(err)
	defer conn.Close()

	cli := network.NewNatpassClient(conn)

	run(cfg, cli)
}
