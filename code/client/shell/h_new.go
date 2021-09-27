package shell

import (
	"fmt"
	"natpass/code/client/pool"
	"net/http"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

// New new shell
func (shell *Shell) New(pool *pool.Pool, w http.ResponseWriter, r *http.Request) {
	id, err := runtime.UUID(16, "0123456789abcdef")
	if err != nil {
		logging.Error("failed to generate link_id for shell: %s, err=%v",
			shell.Name, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	conn := pool.Get(id)
	conn.SendShellCreate(id, shell.cfg)
	conn.AddLink(id)
	logging.Info("new shell: name=%s, id=%s", shell.Name, id)
	fmt.Fprint(w, id)
}
