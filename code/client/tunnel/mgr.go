package tunnel

import "sync"

// Tunnel tunnel interface
type Tunnel interface {
	GetName() string
	GetTypeName() string
	GetTarget() string
}

// Mgr tunnel manager
type Mgr struct {
	sync.RWMutex
	tunnels []Tunnel
}

// New new tunnel manager
func New() *Mgr {
	return &Mgr{}
}

// Add add tunnel
func (mgr *Mgr) Add(tunnel Tunnel) {
	mgr.Lock()
	defer mgr.Unlock()
	mgr.tunnels = append(mgr.tunnels, tunnel)
}
