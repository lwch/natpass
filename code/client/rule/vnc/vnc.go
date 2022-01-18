package vnc

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/jkstack/natpass/code/client/global"
	"github.com/jkstack/natpass/code/client/pool"
	"github.com/jkstack/natpass/code/client/rule"
	"github.com/jkstack/natpass/code/network"
	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

// VNC vnc handler
type VNC struct {
	sync.RWMutex
	Name        string
	cfg         global.Rule
	link        *Link
	chClipboard chan *network.VncClipboard
}

// New new vnc
func New(cfg global.Rule) *VNC {
	return &VNC{
		Name:        cfg.Name,
		cfg:         cfg,
		chClipboard: make(chan *network.VncClipboard),
	}
}

// NewLink new link
func (v *VNC) NewLink(id, remote string, remoteIdx uint32, localConn net.Conn, remoteConn *pool.Conn) rule.Link {
	remoteConn.AddLink(id)
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

// GetName get vnc rule name
func (v *VNC) GetName() string {
	return v.Name
}

// GetTypeName get vnc rule type name
func (v *VNC) GetTypeName() string {
	return "vnc"
}

// GetTarget get target of this rule
func (v *VNC) GetTarget() string {
	return v.cfg.Target
}

// GetLinks get rule links
func (v *VNC) GetLinks() []rule.Link {
	if v.link != nil {
		return []rule.Link{v.link}
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
			logging.Error("close shell: %s, err=%v", v.Name, err)
		}
	}()
	pf := func(cb func(*pool.Pool, http.ResponseWriter, *http.Request)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cb(pl, w, r)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/new", pf(v.New))
	mux.HandleFunc("/ctrl", pf(v.Ctrl))
	mux.HandleFunc("/clipboard", pf(v.Clipboard))
	mux.HandleFunc("/ws/", pf(v.WS))
	mux.HandleFunc("/", v.Render)
	svr := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", v.cfg.LocalAddr, v.cfg.LocalPort),
		Handler: mux,
	}
	runtime.Assert(svr.ListenAndServe())
}
