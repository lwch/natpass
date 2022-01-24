package app

import (
	"github.com/jkstack/natpass/code/client/conn"
	"github.com/jkstack/natpass/code/client/global"
	"github.com/jkstack/natpass/code/client/rule"
	"github.com/jkstack/natpass/code/client/rule/shell"
	"github.com/jkstack/natpass/code/client/rule/vnc"
	"github.com/jkstack/natpass/code/network"
	"github.com/lwch/logging"
)

func (a *App) shellCreate(mgr *rule.Mgr, conn *conn.Conn, msg *network.Msg) {
	create := msg.GetCreq()
	tn := mgr.Get(create.GetName(), msg.GetFrom())
	if tn == nil {
		tn = shell.New(global.Rule{
			Name:   create.GetName(),
			Target: msg.GetFrom(),
			Type:   "shell",
			Exec:   create.GetCshell().GetExec(),
			Env:    create.GetCshell().GetEnv(),
		}, a.cfg.ReadTimeout, a.cfg.WriteTimeout)
		mgr.Add(tn)
	}
	lk := tn.NewLink(msg.GetLinkId(), msg.GetFrom(), nil, conn).(*shell.Link)
	logging.Info("create link %s for shell rule [%s] from %s to %s",
		msg.GetLinkId(), create.GetName(),
		msg.GetFrom(), a.cfg.ID)
	err := lk.Exec()
	if err != nil {
		logging.Error("create shell failed: %v", err)
		conn.SendConnectError(msg.GetFrom(), msg.GetLinkId(), err.Error())
		return
	}
	conn.SendConnectOK(msg.GetFrom(), msg.GetLinkId())
	lk.Forward()
}

func (a *App) vncCreate(confDir string, mgr *rule.Mgr, conn *conn.Conn, msg *network.Msg) {
	create := msg.GetCreq()
	tn := mgr.Get(create.GetName(), msg.GetFrom())
	if tn == nil {
		tn = vnc.New(global.Rule{
			Name:   create.GetName(),
			Target: msg.GetFrom(),
			Type:   "vnc",
			Fps:    create.GetCvnc().GetFps(),
		}, a.cfg.ReadTimeout, a.cfg.WriteTimeout)
		mgr.Add(tn)
	}
	lk := tn.NewLink(msg.GetLinkId(), msg.GetFrom(), nil, conn).(*vnc.Link)
	logging.Info("create link %s for vnc rule [%s] from %s to %s",
		msg.GetLinkId(), create.GetName(),
		msg.GetFrom(), a.cfg.ID)
	lk.SetQuality(create.GetCvnc().GetQuality())
	err := lk.Fork(confDir)
	if err != nil {
		logging.Error("create vnc failed: %v", err)
		conn.SendConnectError(msg.GetFrom(), msg.GetLinkId(), err.Error())
		return
	}
	conn.SendConnectOK(msg.GetFrom(), msg.GetLinkId())
	lk.Forward()
}

func (a *App) benchCreate(confDir string, mgr *rule.Mgr, conn *conn.Conn, msg *network.Msg) {
	create := msg.GetCreq()
	logging.Info("create link %s for bench rule [%s] from %s to %s",
		msg.GetLinkId(), create.GetName(),
		msg.GetFrom(), a.cfg.ID)
	conn.SendConnectOK(msg.GetFrom(), msg.GetLinkId())
}
