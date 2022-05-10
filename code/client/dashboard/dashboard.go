package dashboard

import (
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/lwch/natpass/code/client/conn"
	"github.com/lwch/natpass/code/client/global"
	"github.com/lwch/natpass/code/client/rule"
)

// Dashboard dashboard object
type Dashboard struct {
	cfg     *global.Configure
	conn    *conn.Conn
	mgr     *rule.Mgr
	Version string
}

// New create dashboard object
func New(cfg *global.Configure, conn *conn.Conn, mgr *rule.Mgr, version string) *Dashboard {
	return &Dashboard{
		cfg:     cfg,
		conn:    conn,
		mgr:     mgr,
		Version: version,
	}
}

// ListenAndServe listen and serve http handler
func (db *Dashboard) ListenAndServe(addr string, port uint16) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/info", db.Info)
	mux.HandleFunc("/api/rules", db.Rules)
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/", db.Render)
	svr := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", addr, port),
		Handler: mux,
	}
	return svr.ListenAndServe()
}
