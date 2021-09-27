package shell

import (
	"fmt"
	"natpass/code/client/pool"
	"net/http"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

// New create shell
func (shell *Shell) New(pool *pool.Pool, w http.ResponseWriter, r *http.Request) {
	id, err := runtime.UUID(16, "0123456789abcdef")
	if err != nil {
		logging.Error("failed to generate link_id for shell: %s, err=%v",
			shell.Name, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	conn := pool.Get(id)
	conn.SendShellCreate(shell.cfg.Target, shell.cfg)
	fmt.Fprint(w, id)
}
