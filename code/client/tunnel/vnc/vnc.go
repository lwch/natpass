package vnc

import (
	"natpass/code/client/global"
	"natpass/code/client/pool"
	"natpass/code/client/tunnel"
	"sync"
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
}
