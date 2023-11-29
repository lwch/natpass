package vnc

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/lwch/logging"
	"github.com/lwch/natpass/code/client/conn"
	"github.com/lwch/natpass/code/client/global"
	"github.com/lwch/natpass/code/client/rule"
	"github.com/lwch/natpass/code/network"
	"github.com/lwch/runtime"
)

// VNC vnc handler
type VNC struct {
	sync.RWMutex
	Name         string
	cfg          *global.Rule
	link         *Link
	readTimeout  time.Duration
	writeTimeout time.Duration
	chClipboard  chan *network.VncClipboard
}

// New new vnc
func New(cfg *global.Rule, readTimeout, writeTimeout time.Duration) *VNC {
	return &VNC{
		Name:         cfg.Name,
		cfg:          cfg,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
		chClipboard:  make(chan *network.VncClipboard),
	}
}

// NewLink new link
func (v *VNC) NewLink(id, remote string, localConn net.Conn, remoteConn *conn.Conn) rule.Link {
	remoteConn.AddLink(id)
	link := &Link{
		parent: v,
		id:     id,
		target: remote,
		remote: remoteConn,
	}
	if v.link != nil {
		v.link.Close(true)
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

// OnDisconnect on disconnect message
func (v *VNC) OnDisconnect(id string) {
	// TODO
}

// Handle handle shell
func (v *VNC) Handle(c *conn.Conn) {
	defer func() {
		if err := recover(); err != nil {
			logging.Error("close shell: %s, err=%v", v.Name, err)
		}
	}()
	pf := func(cb func(*conn.Conn, http.ResponseWriter, *http.Request)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cb(c, w, r)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/new", pf(v.New))
	mux.HandleFunc("/ctrl", pf(v.Ctrl))
	mux.HandleFunc("/clipboard", pf(v.Clipboard))
	mux.HandleFunc("/ws/", pf(v.WS))
	mux.HandleFunc("/", v.Render)
	if v.cfg.LocalPort == 0 {
		v.cfg.LocalPort = global.GeneratePort()
		logging.Info("generate port for %s: %d", v.Name, v.cfg.LocalPort)
	}
	svr := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", v.cfg.LocalAddr, v.cfg.LocalPort),
		Handler: mux,
	}
	runtime.Assert(svr.ListenAndServe())
}

func (v *VNC) remove(id string) {
	v.link = nil
}
