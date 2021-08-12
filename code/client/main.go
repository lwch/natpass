package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"natpass/code/client/client"
	"natpass/code/client/global"
	"natpass/code/network"
	"os"

	"github.com/lwch/runtime"
)

func main() {
	conf := flag.String("conf", "", "configure file path")
	flag.Parse()

	if len(*conf) == 0 {
		fmt.Println("missing -conf param")
		os.Exit(1)
	}

	cfg := global.LoadConf(*conf)

	conn, err := tls.Dial("tcp", cfg.Server, nil)
	runtime.Assert(err)
	c := network.NewConn(conn)
	defer c.Close()

	cli := client.New(cfg, c)
	cli.Run()
}
