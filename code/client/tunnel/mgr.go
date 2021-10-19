package tunnel

import (
	"natpass/code/client/pool"
	"net"
	"sync"
)

// Link link interface
type Link interface {
	GetID() string
	// GetBytes rx, tx
	GetBytes() (uint64, uint64)
	// GetPackets rx, tx
	GetPackets() (uint64, uint64)
}

// Tunnel tunnel interface
type Tunnel interface {
	NewLink(id, remote string, remoteIdx uint32, localConn net.Conn, remoteConn *pool.Conn) Link
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

// Get get tunnel by name
func (mgr *Mgr) Get(name, remote string) Tunnel {
	mgr.RLock()
	defer mgr.RUnlock()
	for _, t := range mgr.tunnels {
		if t.GetName() == name && t.GetRemote() == remote {
			return t
		}
	}
	return nil
}

// Range range tunnels
func (mgr *Mgr) Range(fn func(Tunnel)) {
	mgr.RLock()
	defer mgr.RUnlock()
	for _, t := range mgr.tunnels {
		fn(t)
	}
}
