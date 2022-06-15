package rule

import (
	"net"
	"sync"

	"github.com/lwch/natpass/code/client/conn"
)

// Link link interface
type Link interface {
	GetID() string
	// GetBytes rx, tx
	GetBytes() (uint64, uint64)
	// GetPackets rx, tx
	GetPackets() (uint64, uint64)
}

// Rule rule interface
type Rule interface {
	NewLink(id, remote string, localConn net.Conn, remoteConn *conn.Conn) Link
	GetName() string
	GetRemote() string
	GetPort() uint16
	GetTypeName() string
	GetTarget() string
	GetLinks() []Link
}

// Mgr rule manager
type Mgr struct {
	sync.RWMutex
	rules []Rule
}

// New new rule manager
func New() *Mgr {
	return &Mgr{}
}

// Add add rule
func (mgr *Mgr) Add(rule Rule) {
	mgr.Lock()
	defer mgr.Unlock()
	mgr.rules = append(mgr.rules, rule)
}

// Get get rule by name
func (mgr *Mgr) Get(name, remote string) Rule {
	mgr.RLock()
	defer mgr.RUnlock()
	for _, t := range mgr.rules {
		if t.GetName() == name && t.GetRemote() == remote {
			return t
		}
	}
	return nil
}

// Range range rules
func (mgr *Mgr) Range(fn func(Rule)) {
	mgr.RLock()
	defer mgr.RUnlock()
	for _, t := range mgr.rules {
		fn(t)
	}
}
