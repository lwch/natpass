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
	GetName() string
	GetPort() uint16
	GetTypeName() string
}

// LinkedRule linked rule interface
type LinkedRule interface {
	NewLink(id, remote string, localConn net.Conn, remoteConn *conn.Conn) Link
	GetRemote() string
	GetTarget() string
	GetLinks() []Link
	OnDisconnect(string)
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

// GetLinked get rule by name
func (mgr *Mgr) GetLinked(name, remote string) LinkedRule {
	mgr.RLock()
	defer mgr.RUnlock()
	for _, r := range mgr.rules {
		lr, ok := r.(LinkedRule)
		if !ok {
			continue
		}
		if r.GetName() == name && lr.GetRemote() == remote {
			return lr
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

// OnDisconnect on disconnect message
func (mgr *Mgr) OnDisconnect(id string) {
	var links []LinkedRule
	mgr.Range(func(r Rule) {
		if lr, ok := r.(LinkedRule); ok {
			links = append(links, lr)
		}
	})
	for _, link := range links {
		go link.OnDisconnect(id)
	}
}
