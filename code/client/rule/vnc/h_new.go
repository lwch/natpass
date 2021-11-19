package vnc

import (
	"fmt"
	"natpass/code/client/pool"
	"natpass/code/network"
	"net/http"
	"strconv"
	"time"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

// New new vnc
func (v *VNC) New(pool *pool.Pool, w http.ResponseWriter, r *http.Request) {
	if v.link != nil {
		v.link.close()
	}
	q := r.FormValue("quality")
	s := r.FormValue("show_cursor")
	quality, err := strconv.ParseUint(q, 10, 32)
	if err != nil {
		quality = 50
	}
	showCursor, err := strconv.ParseBool(s)
	if err != nil {
		showCursor = false
	}
	id, err := runtime.UUID(16, "0123456789abcdef")
	if err != nil {
		logging.Error("failed to generate link_id for vnc: %s, err=%v",
			v.Name, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	conn := pool.Get(id)
	if v.link != nil {
		old := pool.Get(v.link.id)
		if old != nil {
			old.SendDisconnect(v.link.target, v.link.targetIdx, v.link.id)
		}
	}
	conn.SendConnectVnc(id, v.cfg, quality, showCursor)
	v.link = v.NewLink(id, v.cfg.Target, 0, nil, conn).(*Link)
	ch := conn.ChanRead(id)
	timeout := time.After(conn.ReadTimeout)
	for {
		var msg *network.Msg
		select {
		case msg = <-ch:
		case <-timeout:
			logging.Error("create vnc %s by rule %s failed, timtout", v.link.id, v.link.parent.Name)
			http.Error(w, "timeout", http.StatusBadGateway)
			return
		}
		v.link.SetTargetIdx(msg.GetFromIdx())
		if msg.GetXType() != network.Msg_connect_rep {
			conn.Reset(id, msg)
			time.Sleep(conn.ReadTimeout / 10)
			continue
		}
		rep := msg.GetCrep()
		if !rep.GetOk() {
			logging.Error("create vnc %s by rule %s failed, err=%s",
				v.link.id, v.link.parent.Name, rep.GetMsg())
			http.Error(w, rep.GetMsg(), http.StatusBadGateway)
			return
		}
		break
	}
	logging.Info("new vnc: name=%s, id=%s", v.Name, id)
	fmt.Fprint(w, id)
}
