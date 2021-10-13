package tunnel

import "sync"

type Link interface {
	GetID() string
	// GetBytes rx, tx
	GetBytes() (uint64, uint64)
	// GetPackets rx, tx
	GetPackets() (uint64, uint64)
}

// Tunnel tunnel interface
type Tunnel interface {
	GetName() string
	GetRemote() string
	GetPort() uint16
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

// Range range tunnels
func (mgr *Mgr) Range(fn func(Tunnel)) {
	mgr.RLock()
	defer mgr.RUnlock()
	for _, t := range mgr.tunnels {
		fn(t)
	}
}
