package dashboard

import (
	"fmt"
	"net/http"
)

// Dashboard dashboard object
type Dashboard struct {
}

func New() *Dashboard {
	return &Dashboard{}
}

func (db *Dashboard) ListenAndServe(addr string, port uint16) error {
	mux := http.NewServeMux()
	svr := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", addr, port),
		Handler: mux,
	}
	return svr.ListenAndServe()
}
