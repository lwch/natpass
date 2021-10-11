package main

import (
	"flag"
	"fmt"
	"natpass/code/client/global"
	"natpass/code/client/pool"
	"natpass/code/client/shell"
	"natpass/code/client/tunnel"
	"natpass/code/network"
	"os"
	"path/filepath"
	rt "runtime"
	"time"

	_ "net/http/pprof"

	"github.com/kardianos/service"
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

type app struct {
	cfg *global.Configure
}

func (a *app) Start(s service.Service) error {
	go a.run()
	return nil
}

func (a *app) run() {
	// go func() {
	// 	http.ListenAndServe(":9000", nil)
	// }()

	logging.SetSizeRotate(a.cfg.LogDir, "np-cli", int(a.cfg.LogSize.Bytes()), a.cfg.LogRotate, true)
	defer logging.Flush()

	pl := pool.New(a.cfg)

	for _, t := range a.cfg.Tunnels {
		switch t.Type {
		case "tcp", "udp":
			tn := tunnel.New(t)
			go tn.Handle(pl)
		case "shell":
			sh := shell.New(t)
			go sh.Handle(pl)
		}
	}

	for i := 0; i < a.cfg.Links-pl.Size(); i++ {
		go func() {
			for {
				conn := pl.Get()
				if conn == nil {
					time.Sleep(time.Second)
					continue
				}
				for {
					msg := <-conn.ChanUnknown()
					if msg == nil {
						break
					}
					var linkID string
					switch msg.GetXType() {
					case network.Msg_connect_req:
						switch msg.GetCreq().GetXType() {
						case network.ConnectRequest_tcp, network.ConnectRequest_udp:
							connect(conn, msg)
						case network.ConnectRequest_shell:
							shellCreate(conn, msg)
						}
					default:
						linkID = msg.GetLinkId()
					}
					if len(linkID) > 0 {
						logging.Error("link of %s on connection %d not found, type=%s",
							linkID, conn.Idx, msg.GetXType().String())
						continue
					}
				}
				logging.Info("connection %s-%d exited", a.cfg.ID, conn.Idx)
				time.Sleep(time.Second)
			}
		}()
	}

	select {}
}

func (a *app) Stop(s service.Service) error {
	return nil
}

func main() {
	user := flag.String("user", "", "service user")
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
		Name:         "np-cli",
		DisplayName:  "np-cli",
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
