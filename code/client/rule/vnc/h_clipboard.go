package vnc

import (
	"fmt"
	"net/http"

	"github.com/lwch/natpass/code/client/conn"
)

// Clipboard get/set clipboard
func (v *VNC) Clipboard(conn *conn.Conn, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		v.getClipboard(conn, w, r)
		return
	}
	v.setClipboard(conn, w, r)
}

func (v *VNC) getClipboard(conn *conn.Conn, w http.ResponseWriter, r *http.Request) {
	if v.link == nil {
		http.NotFound(w, r)
		return
	}
	conn.SendVNCClipboardData(v.link.target, v.link.id, false, "")
	data := <-v.chClipboard
	fmt.Fprint(w, data.GetData())
}

func (v *VNC) setClipboard(conn *conn.Conn, w http.ResponseWriter, r *http.Request) {
	data := r.FormValue("data")
	if v.link == nil {
		http.NotFound(w, r)
		return
	}
	conn.SendVNCClipboardData(v.link.target, v.link.id, true, data)
	fmt.Fprint(w, "ok")
}
