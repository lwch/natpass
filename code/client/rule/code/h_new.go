package code

import (
	"fmt"
	"net/http"
	"time"

	"github.com/lwch/logging"
	"github.com/lwch/natpass/code/client/conn"
	"github.com/lwch/natpass/code/network"
	"github.com/lwch/runtime"
)

// New new code-server workspace
func (code *Code) New(conn *conn.Conn, w http.ResponseWriter, r *http.Request) {
	id, err := runtime.UUID(16, "0123456789abcdef")
	if err != nil {
		logging.Error("failed to generate link_id for code-server: %s, err=%v",
			code.Name, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	link := code.NewLink(id, code.cfg.Target, nil, conn).(*Workspace)
	conn.SendConnectReq(id, code.cfg)
	ch := conn.ChanRead(id)
	timeout := time.After(code.readTimeout)
	var repMsg *network.Msg
	for {
		var msg *network.Msg
		select {
		case msg = <-ch:
		case <-timeout:
			logging.Error("create code-server %s by rule %s failed, timtout", link.id, link.parent.Name)
			http.Error(w, "timeout", http.StatusBadGateway)
			return
		}
		if msg.GetXType() != network.Msg_connect_rep {
			conn.Reset(id, msg)
			time.Sleep(code.readTimeout / 10)
			continue
		}
		rep := msg.GetCrep()
		if !rep.GetOk() {
			logging.Error("create code-server %s by rule %s failed, err=%s",
				link.id, link.parent.Name, rep.GetMsg())
			http.Error(w, rep.GetMsg(), http.StatusBadGateway)
			return
		}
		repMsg = msg
		break
	}
	logging.Info("create link %s for code-server rule [%s] from %s to %s",
		link.GetID(), code.cfg.Name,
		repMsg.GetTo(), repMsg.GetFrom())
	fmt.Fprint(w, id)
}
