package shell

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/jkstack/natpass/code/client/pool"
)

// Resize resize terminal
func (shell *Shell) Resize(pool *pool.Pool, w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	rows := r.FormValue("rows")
	cols := r.FormValue("cols")

	shell.RLock()
	link := shell.links[id]
	shell.RUnlock()

	nRows, _ := strconv.ParseUint(rows, 0, 32)
	nCols, _ := strconv.ParseUint(cols, 0, 32)

	link.SendResize(uint32(nRows), uint32(nCols))

	fmt.Fprint(w, "ok")
}
