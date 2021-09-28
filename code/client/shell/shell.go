package shell

import (
	"fmt"
	"natpass/code/client/global"
	"natpass/code/client/pool"
	"net/http"
	"sync"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

type Shell struct {
	sync.RWMutex
	Name  string
	cfg   global.Tunnel
	links map[string]*Link
}

// New new shell
func New(cfg global.Tunnel) *Shell {
	return &Shell{
		Name:  cfg.Name,
		cfg:   cfg,
		links: make(map[string]*Link),
	}
}

// Handle handle shell
func (shell *Shell) Handle(pl *pool.Pool) {
	defer func() {
		if err := recover(); err != nil {
			logging.Error("close shell tunnel: %s, err=%v", shell.Name, err)
		}
	}()
	pf := func(cb func(*pool.Pool, http.ResponseWriter, *http.Request)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cb(pl, w, r)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/new", pf(shell.New))
	mux.HandleFunc("/ws/", pf(shell.WS))
	svr := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", shell.cfg.LocalAddr, shell.cfg.LocalPort),
		Handler: mux,
	}
	runtime.Assert(svr.ListenAndServe())
}
