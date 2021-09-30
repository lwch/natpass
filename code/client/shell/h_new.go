package shell

import (
	"fmt"
	"natpass/code/client/pool"
	"natpass/code/network"
	"net/http"
	"time"

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
	conn.SendShellCreate(id, shell.cfg)
	link := NewLink(shell, id, shell.cfg.Target, conn)
	shell.Lock()
	shell.links[id] = link
	shell.Unlock()
	ch := conn.ChanRead(id)
	timeout := time.After(conn.ReadTimeout)
	for {
		var msg *network.Msg
		select {
		case msg = <-ch:
		case <-timeout:
			logging.Error("create shell %s on tunnel %s failed, timtout", link.id, link.parent.Name)
			http.Error(w, "timeout", http.StatusBadGateway)
			return
		}
		if msg.GetXType() != network.Msg_shell_created {
			conn.Reset(id, msg)
			time.Sleep(conn.ReadTimeout / 10)
			continue
		}
		rep := msg.GetScreated()
		if !rep.GetOk() {
			logging.Error("create shell %s on tunnel %s failed, err=%s",
				link.id, link.parent.Name, msg.GetScreated().GetMsg())
			http.Error(w, rep.GetMsg(), http.StatusBadGateway)
			return
		}
		break
	}
	logging.Info("new shell: name=%s, id=%s", shell.Name, id)
	fmt.Fprint(w, id)
}
