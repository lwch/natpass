package code

import (
	"net/http"

	"github.com/lwch/natpass/code/client/conn"
)

// Forward forward code-server requests
func (code *Code) Forward(conn *conn.Conn, w http.ResponseWriter, r *http.Request) {
}
