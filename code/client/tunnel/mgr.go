package tunnel

import "sync"

type Link interface {
	GetID() string
	// GetBytes tx, rx
	GetBytes() (uint64, uint64)
	// GetPackets tx, rx
	GetPackets() (uint64, uint64)
}

// Tunnel tunnel interface
type Tunnel interface {
	GetName() string
	GetTypeName() string
	GetTarget() string
	GetLinks() []Link
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
