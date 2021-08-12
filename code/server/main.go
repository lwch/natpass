package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"natpass/code/server/global"
	"os"

	"github.com/lwch/daemon"
	"github.com/lwch/logging"
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

	cert, err := tls.LoadX509KeyPair(cfg.TLSCrt, cfg.TLSKey)
	runtime.Assert(err)
	l, err := tls.Listen("tcp", fmt.Sprintf(":%d", cfg.Listen), &tls.Config{
		Certificates: []tls.Certificate{cert},
	})
	runtime.Assert(err)
	logging.Info("listen on %d", cfg.Listen)

	run(cfg, l)
}
