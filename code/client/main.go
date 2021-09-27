package main

import (
	"flag"
	"fmt"
	"natpass/code/client/global"
	"natpass/code/client/pool"
	"natpass/code/client/shell"
	"natpass/code/client/tunnel"
	"natpass/code/network"
	"net"
	"os"
	"path/filepath"
	rt "runtime"
	"strconv"
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
						connect(pl, conn, msg.GetLinkId(), msg.GetFrom(), msg.GetTo(),
							msg.GetFromIdx(), msg.GetToIdx(), msg.GetCreq())
					case network.Msg_shell_create:
					case network.Msg_connect_rep,
						network.Msg_disconnect,
						network.Msg_forward:
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

func connect(pool *pool.Pool, conn *pool.Conn, id, from, to string, fromIdx, toIdx uint32, req *network.ConnectRequest) {
	dial := "tcp"
	if req.GetXType() == network.ConnectRequest_udp {
		dial = "udp"
	}
	link, err := net.Dial(dial, fmt.Sprintf("%s:%d", req.GetAddr(), req.GetPort()))
	if err != nil {
		conn.SendConnectError(from, fromIdx, id, err.Error())
		return
	}
	host, pt, _ := net.SplitHostPort(link.LocalAddr().String())
	port, _ := strconv.ParseUint(pt, 10, 16)
	tn := tunnel.New(global.Tunnel{
		Name:       req.GetName(),
		Target:     from,
		Type:       dial,
		LocalAddr:  host,
		LocalPort:  uint16(port),
		RemoteAddr: req.GetAddr(),
		RemotePort: uint16(req.GetPort()),
	})
	lk := tunnel.NewLink(tn, id, from, link, conn)
	lk.SetTargetIdx(fromIdx)
	conn.SendConnectOK(from, fromIdx, id)
	lk.Forward()
	lk.OnWork <- struct{}{}
}
