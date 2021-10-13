package dashboard

import (
	"fmt"
	"natpass/code/client/global"
	"natpass/code/client/pool"
	"natpass/code/client/tunnel"
	"net/http"
)

// Dashboard dashboard object
type Dashboard struct {
	cfg     *global.Configure
	pl      *pool.Pool
	mgr     *tunnel.Mgr
	Version string
}

func New(cfg *global.Configure, pl *pool.Pool, mgr *tunnel.Mgr, version string) *Dashboard {
	return &Dashboard{
		cfg:     cfg,
		pl:      pl,
		mgr:     mgr,
		Version: version,
	}
}

func (db *Dashboard) ListenAndServe(addr string, port uint16) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/info", db.Info)
	mux.HandleFunc("/api/tunnels", db.Tunnels)
	mux.HandleFunc("/", db.Render)
	svr := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", addr, port),
		Handler: mux,
	}
	return svr.ListenAndServe()
}
