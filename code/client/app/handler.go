package app

import (
	"github.com/lwch/logging"
	"github.com/lwch/natpass/code/client/conn"
	"github.com/lwch/natpass/code/client/global"
	"github.com/lwch/natpass/code/client/rule"
	"github.com/lwch/natpass/code/client/rule/code"
	"github.com/lwch/natpass/code/client/rule/shell"
	"github.com/lwch/natpass/code/client/rule/vnc"
	"github.com/lwch/natpass/code/network"
)

/*
this is the handler function file like http.HandlerFunc.

for linked rule:
1. get rule from manager
2. create rule if not exists by its type
3. create link call NewLink
4. do the initialize logic for the link
5. response connect ok message
6. loop forward

no linked rule:
TODO
*/

func (p *program) shellCreate(mgr *rule.Mgr, conn *conn.Conn, msg *network.Msg) {
	create := msg.GetCreq()
	tn := mgr.GetLinked(create.GetName(), msg.GetFrom())
	if tn == nil {
		tn = shell.New(&global.Rule{
			Name:   create.GetName(),
			Target: msg.GetFrom(),
			Type:   "shell",
			Exec:   create.GetCshell().GetExec(),
			Env:    create.GetCshell().GetEnv(),
		}, p.cfg.ReadTimeout, p.cfg.WriteTimeout)
		mgr.Add(tn.(rule.Rule))
	}
	lk := tn.NewLink(msg.GetLinkId(), msg.GetFrom(), nil, conn).(*shell.Link)
	logging.Info("create link %s for shell rule [%s] from %s to %s",
		msg.GetLinkId(), create.GetName(),
		msg.GetFrom(), p.cfg.ID)
	err := lk.Exec()
	if err != nil {
		logging.Error("create shell failed: %v", err)
		conn.SendConnectError(msg.GetFrom(), msg.GetLinkId(), err.Error())
		return
	}
	conn.SendConnectOK(msg.GetFrom(), msg.GetLinkId())
	lk.Forward()
}

func (p *program) vncCreate(confDir string, mgr *rule.Mgr, conn *conn.Conn, msg *network.Msg) {
	create := msg.GetCreq()
	tn := mgr.GetLinked(create.GetName(), msg.GetFrom())
	if tn == nil {
		tn = vnc.New(&global.Rule{
			Name:   create.GetName(),
			Target: msg.GetFrom(),
			Type:   "vnc",
			Fps:    create.GetCvnc().GetFps(),
		}, p.cfg.ReadTimeout, p.cfg.WriteTimeout)
		mgr.Add(tn.(rule.Rule))
	}
	lk := tn.NewLink(msg.GetLinkId(), msg.GetFrom(), nil, conn).(*vnc.Link)
	logging.Info("create link %s for vnc rule [%s] from %s to %s",
		msg.GetLinkId(), create.GetName(),
		msg.GetFrom(), p.cfg.ID)
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

func (p *program) benchCreate(confDir string, mgr *rule.Mgr, conn *conn.Conn, msg *network.Msg) {
	create := msg.GetCreq()
	logging.Info("create link %s for bench rule [%s] from %s to %s",
		msg.GetLinkId(), create.GetName(),
		msg.GetFrom(), p.cfg.ID)
	conn.SendConnectOK(msg.GetFrom(), msg.GetLinkId())
}

func (p *program) codeCreate(confDir string, mgr *rule.Mgr, conn *conn.Conn, msg *network.Msg) {
	create := msg.GetCreq()
	tn := mgr.GetLinked(create.GetName(), msg.GetFrom())
	if tn == nil {
		tn = code.New(&global.Rule{
			Name:   create.GetName(),
			Target: msg.GetFrom(),
			Type:   "code-server",
		}, p.cfg.ReadTimeout, p.cfg.WriteTimeout)
		mgr.Add(tn.(rule.Rule))
	}
	workspace := tn.NewLink(msg.GetLinkId(), msg.GetFrom(), nil, conn).(*code.Workspace)
	logging.Info("create link %s for code-server rule [%s] from %s to %s",
		msg.GetLinkId(), create.GetName(),
		msg.GetFrom(), p.cfg.ID)
	err := workspace.Exec(p.cfg.CodeDir)
	if err != nil {
		logging.Error("create vnc failed: %v", err)
		conn.SendConnectError(msg.GetFrom(), msg.GetLinkId(), err.Error())
		return
	}
	conn.SendConnectOK(msg.GetFrom(), msg.GetLinkId())
	workspace.Forward()
}
