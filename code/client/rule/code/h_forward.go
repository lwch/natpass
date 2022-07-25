package code

import (
	"net/http"
	"strings"

	"github.com/lwch/natpass/code/client/conn"
)

// Forward forward code-server requests
func (code *Code) Forward(conn *conn.Conn, w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/forward/")
	id = id[:strings.Index(id, "/")]

	code.RLock()
	workspace := code.workspace[id]
	code.RUnlock()

	if workspace == nil {
		http.NotFound(w, r)
		return
	}

	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/forward/"+id)
	if len(r.URL.Path) == 0 {
		r.URL.Path = "/"
	}

	if code.isWebsocket(r) {
		code.handleWebsocket(workspace, w, r)
	} else {
		code.handleRequest(workspace, w, r)
	}
}

func (code *Code) isWebsocket(r *http.Request) bool {
	upgrade := r.Header.Get("Connection")
	return upgrade == "Upgrade"
}
