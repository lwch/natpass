package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"natpass/code/client/client"
	"natpass/code/client/global"
	"natpass/code/network"
	"os"

	"github.com/lwch/daemon"
	"github.com/lwch/runtime"
)

func main() {
	bak := flag.Bool("d", false, "backend running")
	pid := flag.String("pid", "", "pid file dir")
	user := flag.String("u", "", "daemon user")
	conf := flag.String("conf", "", "configure file path")
	flag.Parse()

	if len(*conf) == 0 {
		fmt.Println("missing -conf param")
		os.Exit(1)
	}

	if *bak {
		daemon.Start(0, *pid, *user, "-conf", *conf)
		return
	}

	cfg := global.LoadConf(*conf)

	conn, err := tls.Dial("tcp", cfg.Server, nil)
	runtime.Assert(err)
	c := network.NewConn(conn)
	defer c.Close()

	cli := client.New(cfg, c)
	cli.Run()
}
