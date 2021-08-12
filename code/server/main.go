package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"natpass/code/server/global"
	"os"

	"github.com/lwch/logging"
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

	cert, err := tls.LoadX509KeyPair(cfg.TLSCrt, cfg.TLSKey)
	runtime.Assert(err)
	l, err := tls.Listen("tcp", fmt.Sprintf(":%d", cfg.Listen), &tls.Config{
		Certificates: []tls.Certificate{cert},
	})
	runtime.Assert(err)
	logging.Info("listen on %d", cfg.Listen)

	run(cfg, l)
}
