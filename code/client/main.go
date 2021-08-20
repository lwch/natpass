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

	logging.SetSizeRotate(cfg.LogDir, "np-cli", int(cfg.LogSize.Bytes()), cfg.LogRotate, true)
	defer logging.Flush()

	conn, err := tls.Dial("tcp", cfg.Server, nil)
	runtime.Assert(err)
	c := network.NewConn(conn)
	defer c.Close()

	cli := client.New(cfg, c)
	cli.Run()
}
