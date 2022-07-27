package code

import (
	"net/http"

	"github.com/lwch/natpass/code/client/conn"
)

// New new code-server workspace
func (code *Code) New(conn *conn.Conn, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(code.cfg.Name))
}
