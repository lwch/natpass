package vnc

import (
	"fmt"
	"natpass/code/client/global"
	"natpass/code/client/pool"
	"natpass/code/client/tunnel"
	"net"
	"net/http"
	"sync"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

// VNC vnc handler
type VNC struct {
	sync.RWMutex
	Name string
	cfg  global.Tunnel
	link *Link
}

// New new vnc
func New(cfg global.Tunnel) *VNC {
	return &VNC{
		Name: cfg.Name,
		cfg:  cfg,
	}
}

// NewLink new link
func (v *VNC) NewLink(id, remote string, remoteIdx uint32, localConn net.Conn, remoteConn *pool.Conn) tunnel.Link {
	remoteConn.AddLink(id)
	logging.Info("create link %s for tunnel %s on connection %d",
		id, v.Name, remoteConn.Idx)
	link := &Link{
		parent:    v,
		id:        id,
		target:    remote,
		targetIdx: remoteIdx,
		remote:    remoteConn,
	}
	if v.link != nil {
		v.link.close()
	}
	v.link = link
	return link
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
	if v.link != nil {
		return []tunnel.Link{v.link}
	}
	return nil
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
	mux.HandleFunc("/ws/", pf(v.WS))
	mux.HandleFunc("/", v.Render)
	svr := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", v.cfg.LocalAddr, v.cfg.LocalPort),
		Handler: mux,
	}
	runtime.Assert(svr.ListenAndServe())
}
