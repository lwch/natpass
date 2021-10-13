package dashboard

import (
	"fmt"
	"net/http"
)

// Dashboard dashboard object
type Dashboard struct {
	Version string
}

func New(version string) *Dashboard {
	return &Dashboard{
		Version: version,
	}
}

func (db *Dashboard) ListenAndServe(addr string, port uint16) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", db.Render)
	svr := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", addr, port),
		Handler: mux,
	}
	return svr.ListenAndServe()
}
