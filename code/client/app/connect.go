package app

import (
	"github.com/jkstack/natpass/code/client/global"
	"github.com/jkstack/natpass/code/client/pool"
	"github.com/jkstack/natpass/code/client/rule"
	"github.com/jkstack/natpass/code/client/rule/shell"
	"github.com/jkstack/natpass/code/client/rule/vnc"
	"github.com/jkstack/natpass/code/network"
	"github.com/lwch/logging"
)

func (a *App) shellCreate(mgr *rule.Mgr, conn *pool.Conn, msg *network.Msg) {
	create := msg.GetCreq()
	tn := mgr.Get(create.GetName(), msg.GetFrom())
	if tn == nil {
		tn = shell.New(global.Rule{
			Name:   create.GetName(),
			Target: msg.GetFrom(),
			Type:   "shell",
			Exec:   create.GetCshell().GetExec(),
			Env:    create.GetCshell().GetEnv(),
		})
		mgr.Add(tn)
	}
	lk := tn.NewLink(msg.GetLinkId(), msg.GetFrom(), msg.GetFromIdx(), nil, conn).(*shell.Link)
	logging.Info("create link %s for shell rule [%s] from %s-%d to %s-%d",
		msg.GetLinkId(), create.GetName(),
		msg.GetFrom(), msg.GetFromIdx(),
		a.cfg.ID, conn.GetIdx())
	err := lk.Exec()
	if err != nil {
		logging.Error("create shell failed: %v", err)
		conn.SendConnectError(msg.GetFrom(), msg.GetFromIdx(), msg.GetLinkId(), err.Error())
		return
	}
	conn.SendConnectOK(msg.GetFrom(), msg.GetFromIdx(), msg.GetLinkId())
	lk.Forward()
}

func (a *App) vncCreate(confDir string, mgr *rule.Mgr, conn *pool.Conn, msg *network.Msg) {
	create := msg.GetCreq()
	tn := mgr.Get(create.GetName(), msg.GetFrom())
	if tn == nil {
		tn = vnc.New(global.Rule{
			Name:   create.GetName(),
			Target: msg.GetFrom(),
			Type:   "vnc",
			Fps:    create.GetCvnc().GetFps(),
		})
		mgr.Add(tn)
	}
	lk := tn.NewLink(msg.GetLinkId(), msg.GetFrom(), msg.GetFromIdx(), nil, conn).(*vnc.Link)
	logging.Info("create link %s for vnc rule [%s] from %s-%d to %s-%d",
		msg.GetLinkId(), create.GetName(),
		msg.GetFrom(), msg.GetFromIdx(),
		a.cfg.ID, conn.GetIdx())
	lk.SetQuality(create.GetCvnc().GetQuality())
	err := lk.Fork(confDir)
	if err != nil {
		logging.Error("create vnc failed: %v", err)
		conn.SendConnectError(msg.GetFrom(), msg.GetFromIdx(), msg.GetLinkId(), err.Error())
		return
	}
	conn.SendConnectOK(msg.GetFrom(), msg.GetFromIdx(), msg.GetLinkId())
	lk.Forward()
}
