package vnc

import (
	"fmt"
	"natpass/code/client/pool"
	"net/http"
	"strconv"
)

// Ctrl change vnc tunnel config
func (v *VNC) Ctrl(pool *pool.Pool, w http.ResponseWriter, r *http.Request) {
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
	if v.link == nil {
		http.NotFound(w, r)
		return
	}
	conn := pool.Get(v.link.id)
	conn.SendVNCCtrl(v.link.target, v.link.targetIdx, v.link.id, quality, showCursor)
	fmt.Fprint(w, "ok")
}
