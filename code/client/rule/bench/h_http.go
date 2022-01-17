package bench

import (
	"fmt"
	"net/http"

	"github.com/jkstack/natpass/code/client/pool"
	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

func (bench *Bench) http(pool *pool.Pool, w http.ResponseWriter, r *http.Request) {
	id, err := runtime.UUID(16, "0123456789abcdef")
	if err != nil {
		logging.Error("failed to generate link_id for bench: %s, err=%v",
			bench.Name, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	conn := pool.Get(id)
	conn.AddLink(id)
	conn.SendConnectReq(id, bench.cfg)
	ch := conn.ChanRead(id)
	<-ch
	conn.RemoveLink(id)
	fmt.Fprint(w, id)
}
