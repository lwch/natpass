package vnc

import (
	"natpass/code/client/pool"
	"net/http"
	"strings"
)

func (v *VNC) WS(pool *pool.Pool, w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/ws/")
	conn := pool.Get(id)
	if conn == nil {
		http.NotFound(w, r)
		return
	}
	ch := conn.ChanRead(id)
	for {
		msg := <-ch
		// logging.Info("%s", msg.String())
		_ = msg
	}
}
