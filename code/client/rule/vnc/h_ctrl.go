package vnc

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/lwch/natpass/code/client/conn"
)

// Ctrl change vnc rule config
func (v *VNC) Ctrl(conn *conn.Conn, w http.ResponseWriter, r *http.Request) {
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
	conn.SendVNCCtrl(v.link.target, v.link.id, quality, showCursor)
	fmt.Fprint(w, "ok")
}
