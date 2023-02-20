package app

import (
	"os"
	rt "runtime"

	"github.com/common-nighthawk/go-figure"
	"github.com/kardianos/service"
	"github.com/lwch/logging"
	"github.com/lwch/natpass/code/client/conn"
	"github.com/lwch/natpass/code/client/dashboard"
	"github.com/lwch/natpass/code/client/global"
	"github.com/lwch/natpass/code/client/rule"
	"github.com/lwch/natpass/code/client/rule/bench"
	"github.com/lwch/natpass/code/client/rule/code"
	"github.com/lwch/natpass/code/client/rule/shell"
	"github.com/lwch/natpass/code/client/rule/vnc"
	"github.com/lwch/natpass/code/network"
	"github.com/lwch/runtime"
)

type program struct {
	confDir string
	cfg     *global.Configure
	conn    *conn.Conn
}

func newProgram() *program {
	return &program{}
}

func (p *program) setConfDir(dir string) *program {
	p.confDir = dir
	return p
}

func (p *program) setConfigure(cfg *global.Configure) *program {
	p.cfg = cfg
	return p
}

// Start main entry for service
func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

// Stop stop service callback
func (*program) Stop(s service.Service) error {
	return nil
}

func (p *program) run() {
	// go func() {
	// 	http.ListenAndServe(":9000", nil)
	// }()

	// initialize logging
	stdout := true
	if rt.GOOS == "windows" {
		stdout = false
	}
	logging.SetSizeRotate(logging.SizeRotateConfig{
		Dir:         p.cfg.LogDir,
		Name:        "np-cli",
		Size:        int64(p.cfg.LogSize.Bytes()),
		Rotate:      p.cfg.LogRotate,
		WriteStdout: stdout,
		WriteFile:   true,
	})
	defer logging.Flush()

	fg := figure.NewFigure("NatPass", "alligator2", false)
	figure.Write(&logging.DefaultLogger, fg)
	logging.DefaultLogger.Write(nil)

	// create connection and handshake
	p.conn = conn.New(p.cfg)

	// build rule manager
	mgr := rule.New()

	// add rules from configure file, wait http request from web browser
	for _, t := range p.cfg.Rules {
		switch t.Type {
		case "shell":
			sh := shell.New(t, p.cfg.ReadTimeout, p.cfg.WriteTimeout)
			mgr.Add(sh)
			go sh.Handle(p.conn)
		case "vnc":
			v := vnc.New(t, p.cfg.ReadTimeout, p.cfg.WriteTimeout)
			mgr.Add(v)
			go v.Handle(p.conn)
		case "bench":
			b := bench.New(t)
			mgr.Add(b)
			go b.Handle(p.conn)
		case "code-server":
			cs := code.New(t, p.cfg.ReadTimeout, p.cfg.WriteTimeout)
			mgr.Add(cs)
			go cs.Handle(p.conn)
		}
	}

	// handle request from remote node
	go func() {
		for {
			msg := <-p.conn.ChanUnknown()
			var linkID string
			switch msg.GetXType() {
			case network.Msg_connect_req:
				switch msg.GetCreq().GetXType() {
				case network.ConnectRequest_shell:
					// fork /bin/bash command and ack
					p.shellCreate(mgr, p.conn, msg)
				case network.ConnectRequest_vnc:
					// fork np-cli vnc child process and ack
					p.vncCreate(p.confDir, mgr, p.conn, msg)
				case network.ConnectRequest_bench:
					// bench handler response ok directly
					p.benchCreate(p.confDir, mgr, p.conn, msg)
				case network.ConnectRequest_code:
					// fork code-server child process and ack
					p.codeCreate(p.confDir, mgr, p.conn, msg)
				}
			default:
				linkID = msg.GetLinkId()
			}
			// invalid message type, close channel directly
			if len(linkID) > 0 {
				p.conn.ChanClose(linkID)
				logging.Error("link of %s not found, type=%s",
					linkID, msg.GetXType().String())
				continue
			}
		}
	}()

	// on disconnect message dispatcher, also to close the forked process
	go func() {
		for {
			id := <-p.conn.ChanDisconnect()
			mgr.OnDisconnect(id)
		}
	}()

	if p.cfg.DashboardEnabled {
		// if the dashboard is enabled, wait the connection close async
		go func() {
			p.conn.Wait()
			logging.Flush()
			os.Exit(1)
		}()
		// handle dashboard
		db := dashboard.New(p.cfg, p.conn, mgr, Version)
		runtime.Assert(db.ListenAndServe(p.cfg.DashboardListen, p.cfg.DashboardPort))
	} else {
		// wait the connection close
		p.conn.Wait()
	}
}
