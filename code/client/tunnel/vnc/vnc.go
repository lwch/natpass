package vnc

import (
	"fmt"
	"natpass/code/client/global"
	"natpass/code/client/pool"
	"natpass/code/client/tunnel"
	"net/http"
	"sync"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

// VNC vnc handler
type VNC struct {
	sync.RWMutex
	Name  string
	cfg   global.Tunnel
	links map[string]*Link
}

// New new vnc
func New(cfg global.Tunnel) *VNC {
	return &VNC{
		Name:  cfg.Name,
		cfg:   cfg,
		links: make(map[string]*Link),
	}
}

// GetName get vnc tunnel name
func (v *VNC) GetName() string {
	return v.Name
}

// GetTypeName get vnc tunnel type name
func (v *VNC) GetTypeName() string {
	return "vnc"
}

// GetTarget get target of this tunnel
func (v *VNC) GetTarget() string {
	return v.cfg.Target
}

// GetLinks get tunnel links
func (v *VNC) GetLinks() []tunnel.Link {
	ret := make([]tunnel.Link, 0, len(v.links))
	v.RLock()
	for _, link := range v.links {
		ret = append(ret, link)
	}
	v.RUnlock()
	return ret
}

// GetRemote get remote target name
func (v *VNC) GetRemote() string {
	return v.cfg.Target
}

// GetPort get listen port
func (v *VNC) GetPort() uint16 {
	return v.cfg.LocalPort
}

// Handle handle shell
func (v *VNC) Handle(pl *pool.Pool) {
	defer func() {
		if err := recover(); err != nil {
			logging.Error("close shell tunnel: %s, err=%v", v.Name, err)
		}
	}()
	pf := func(cb func(*pool.Pool, http.ResponseWriter, *http.Request)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cb(pl, w, r)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/new", pf(v.New))
	mux.HandleFunc("/", v.Render)
	svr := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", v.cfg.LocalAddr, v.cfg.LocalPort),
		Handler: mux,
	}
	runtime.Assert(svr.ListenAndServe())
}
