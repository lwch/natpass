package shell

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jkstack/natpass/code/client/pool"
	"github.com/jkstack/natpass/code/network"
	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

// New new shell
func (shell *Shell) New(pool *pool.Pool, w http.ResponseWriter, r *http.Request) {
	id, err := runtime.UUID(16, "0123456789abcdef")
	if err != nil {
		logging.Error("failed to generate link_id for shell: %s, err=%v",
			shell.Name, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	conn := pool.Get(id)
	link := shell.NewLink(id, shell.cfg.Target, 0, nil, conn).(*Link)
	conn.SendConnectReq(id, shell.cfg)
	ch := conn.ChanRead(id)
	timeout := time.After(conn.ReadTimeout)
	for {
		var msg *network.Msg
		select {
		case msg = <-ch:
		case <-timeout:
			logging.Error("create shell %s by rule %s failed, timtout", link.id, link.parent.Name)
			http.Error(w, "timeout", http.StatusBadGateway)
			return
		}
		link.SetTargetIdx(msg.GetFromIdx())
		if msg.GetXType() != network.Msg_connect_rep {
			conn.Reset(id, msg)
			time.Sleep(conn.ReadTimeout / 10)
			continue
		}
		rep := msg.GetCrep()
		if !rep.GetOk() {
			logging.Error("create shell %s by rule %s failed, err=%s",
				link.id, link.parent.Name, rep.GetMsg())
			http.Error(w, rep.GetMsg(), http.StatusBadGateway)
			return
		}
		break
	}
	logging.Info("new shell: name=%s, id=%s", shell.Name, id)
	fmt.Fprint(w, id)
}
