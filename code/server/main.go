package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"natpass/code/server/global"
	"natpass/code/server/handler"
	"net"
	"os"
	"path/filepath"
	rt "runtime"

	"github.com/kardianos/service"
	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

var (
	version      string = "0.0.0"
	gitHash      string
	gitReversion string
	buildTime    string
)

func showVersion() {
	fmt.Printf("version: v%s\ntime: %s\ncommit: %s.%s\n",
		version,
		buildTime,
		gitHash, gitReversion)
	os.Exit(0)
}

type app struct {
	cfg *global.Configure
}

func (a *app) Start(s service.Service) error {
	go a.run()
	return nil
}

func (a *app) run() {
	logging.SetSizeRotate(a.cfg.LogDir, "np-svr", int(a.cfg.LogSize.Bytes()), a.cfg.LogRotate, true)
	defer logging.Flush()

	cert, err := tls.LoadX509KeyPair(a.cfg.TLSCrt, a.cfg.TLSKey)
	runtime.Assert(err)
	l, err := tls.Listen("tcp", fmt.Sprintf(":%d", a.cfg.Listen), &tls.Config{
		Certificates: []tls.Certificate{cert},
	})
	runtime.Assert(err)
	logging.Info("listen on %d", a.cfg.Listen)

	run(a.cfg, l)
}

func (a *app) Stop(s service.Service) error {
	return nil
}

func main() {
	user := flag.String("user", "", "daemon user")
	conf := flag.String("conf", "", "configure file path")
	version := flag.Bool("version", false, "show version info")
	act := flag.String("action", "", "install or uninstall")
	flag.Parse()

	if *version {
		showVersion()
		os.Exit(0)
	}

	if len(*conf) == 0 {
		fmt.Println("missing -conf param")
		os.Exit(1)
	}

	dir, err := filepath.Abs(*conf)
	runtime.Assert(err)

	var depends []string
	if rt.GOOS != "windows" {
		depends = append(depends, "After=network.target")
	}

	appCfg := &service.Config{
		Name:         "np-svr",
		DisplayName:  "np-svr",
		Description:  "nat forward service",
		UserName:     *user,
		Arguments:    []string{"-conf", dir},
		Dependencies: depends,
	}

	cfg := global.LoadConf(*conf)

	app := &app{cfg: cfg}
	sv, err := service.New(app, appCfg)
	runtime.Assert(err)

	switch *act {
	case "install":
		runtime.Assert(sv.Install())
	case "uninstall":
		runtime.Assert(sv.Uninstall())
	default:
		runtime.Assert(sv.Run())
	}
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
