package vnc

import (
	"fmt"
	"natpass/code/client/pool"
	"net/http"
)

// Clipboard get/set clipboard
func (v *VNC) Clipboard(pool *pool.Pool, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		v.getClipboard(pool, w, r)
		return
	}
	v.setClipboard(pool, w, r)
}

func (v *VNC) getClipboard(pool *pool.Pool, w http.ResponseWriter, r *http.Request) {
}

func (v *VNC) setClipboard(pool *pool.Pool, w http.ResponseWriter, r *http.Request) {
	data := r.FormValue("data")
	if v.link == nil {
		http.NotFound(w, r)
		return
	}
	conn := pool.Get(v.link.id)
	conn.SendVNCClipboardSet(v.link.target, v.link.targetIdx, v.link.id, data)
	fmt.Fprint(w, "ok")
}
