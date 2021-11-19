package app

import (
	"natpass/code/client/dashboard"
	"natpass/code/client/global"
	"natpass/code/client/pool"
	"natpass/code/client/rule"
	"natpass/code/client/rule/shell"
	"natpass/code/client/rule/vnc"
	"natpass/code/network"
	rt "runtime"
	"time"

	"github.com/kardianos/service"
	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

// App application
type App struct {
	confDir string
	cfg     *global.Configure
	version string
}

// New create application
func New(ver, dir string, cfg *global.Configure) *App {
	return &App{version: ver, confDir: dir, cfg: cfg}
}

// Start start application
func (a *App) Start(s service.Service) error {
	go a.run()
	return nil
}

// Stop stop application
func (a *App) Stop(s service.Service) error {
	return nil
}

func (a *App) run() {
	// go func() {
	// 	http.ListenAndServe(":9000", nil)
	// }()

	stdout := true
	if rt.GOOS == "windows" {
		stdout = false
	}
	logging.SetSizeRotate(a.cfg.LogDir, "np-cli", int(a.cfg.LogSize.Bytes()), a.cfg.LogRotate, stdout)
	defer logging.Flush()

	pl := pool.New(a.cfg)
	mgr := rule.New()

	for _, t := range a.cfg.Rules {
		switch t.Type {
		case "shell":
			sh := shell.New(t)
			mgr.Add(sh)
			go sh.Handle(pl)
		case "vnc":
			v := vnc.New(t)
			mgr.Add(v)
			go v.Handle(pl)
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
						case network.ConnectRequest_shell:
							a.shellCreate(mgr, conn, msg)
						case network.ConnectRequest_vnc:
							a.vncCreate(a.confDir, mgr, conn, msg)
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

	if a.cfg.DashboardEnabled {
		db := dashboard.New(a.cfg, pl, mgr, a.version)
		runtime.Assert(db.ListenAndServe(a.cfg.DashboardListen, a.cfg.DashboardPort))
	} else {
		select {}
	}
}
