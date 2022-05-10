package app

import (
	rt "runtime"

	"github.com/kardianos/service"
	"github.com/lwch/logging"
	"github.com/lwch/natpass/code/client/conn"
	"github.com/lwch/natpass/code/client/dashboard"
	"github.com/lwch/natpass/code/client/global"
	"github.com/lwch/natpass/code/client/rule"
	"github.com/lwch/natpass/code/client/rule/bench"
	"github.com/lwch/natpass/code/client/rule/shell"
	"github.com/lwch/natpass/code/client/rule/vnc"
	"github.com/lwch/natpass/code/network"
	"github.com/lwch/runtime"
)

// App application
type App struct {
	confDir string
	cfg     *global.Configure
	version string
	conn    *conn.Conn
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
	logging.SetSizeRotate(logging.SizeRotateConfig{
		Dir:         a.cfg.LogDir,
		Name:        "np-cli",
		Size:        int64(a.cfg.LogSize.Bytes()),
		Rotate:      a.cfg.LogRotate,
		WriteStdout: stdout,
		WriteFile:   true,
	})
	defer logging.Flush()

	a.conn = conn.New(a.cfg)
	mgr := rule.New()

	for _, t := range a.cfg.Rules {
		switch t.Type {
		case "shell":
			sh := shell.New(t, a.cfg.ReadTimeout, a.cfg.WriteTimeout)
			mgr.Add(sh)
			go sh.Handle(a.conn)
		case "vnc":
			v := vnc.New(t, a.cfg.ReadTimeout, a.cfg.WriteTimeout)
			mgr.Add(v)
			go v.Handle(a.conn)
		case "bench":
			b := bench.New(t)
			mgr.Add(b)
			go b.Handle(a.conn)
		}
	}

	go func() {
		for {
			msg := <-a.conn.ChanUnknown()
			var linkID string
			switch msg.GetXType() {
			case network.Msg_connect_req:
				switch msg.GetCreq().GetXType() {
				case network.ConnectRequest_shell:
					a.shellCreate(mgr, a.conn, msg)
				case network.ConnectRequest_vnc:
					a.vncCreate(a.confDir, mgr, a.conn, msg)
				case network.ConnectRequest_bench:
					a.benchCreate(a.confDir, mgr, a.conn, msg)
				}
			default:
				linkID = msg.GetLinkId()
			}
			if len(linkID) > 0 {
				logging.Error("link of %s not found, type=%s",
					linkID, msg.GetXType().String())
				continue
			}
		}
	}()

	if a.cfg.DashboardEnabled {
		db := dashboard.New(a.cfg, a.conn, mgr, a.version)
		runtime.Assert(db.ListenAndServe(a.cfg.DashboardListen, a.cfg.DashboardPort))
	} else {
		select {}
	}
}
