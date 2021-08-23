package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"natpass/code/server/global"
	"natpass/code/server/handler"
	"net"
	"os"

	"github.com/lwch/daemon"
	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

var (
	_VERSION       string = "0.0.0"
	_GIT_HASH      string
	_GIT_REVERSION string
	_BUILD_TIME    string
)

func showVersion() {
	fmt.Printf("version: v%s\ntime: %s\ncommit: %s.%s\n",
		_VERSION,
		_BUILD_TIME,
		_GIT_HASH, _GIT_REVERSION)
	os.Exit(0)
}

func main() {
	bak := flag.Bool("d", false, "backend running")
	pid := flag.String("pid", "", "pid file dir")
	user := flag.String("u", "", "daemon user")
	conf := flag.String("conf", "", "configure file path")
	version := flag.Bool("v", false, "show version info")
	flag.Parse()

	if *version {
		showVersion()
		os.Exit(0)
	}

	if len(*conf) == 0 {
		fmt.Println("missing -conf param")
		os.Exit(1)
	}

	if *bak {
		daemon.Start(0, *pid, *user, "-conf", *conf)
		return
	}

	cfg := global.LoadConf(*conf)

	logging.SetSizeRotate(cfg.LogDir, "np-svr", int(cfg.LogSize.Bytes()), cfg.LogRotate, true)
	defer logging.Flush()

	cert, err := tls.LoadX509KeyPair(cfg.TLSCrt, cfg.TLSKey)
	runtime.Assert(err)
	l, err := tls.Listen("tcp", fmt.Sprintf(":%d", cfg.Listen), &tls.Config{
		Certificates: []tls.Certificate{cert},
	})
	runtime.Assert(err)
	logging.Info("listen on %d", cfg.Listen)

	run(cfg, l)
}

func run(cfg *global.Configure, l net.Listener) {
	h := handler.New(cfg)
	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		go h.Handle(conn)
	}
}
