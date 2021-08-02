package tunnel

import "sync"

type Mgr struct {
	sync.RWMutex
	data map[string]*Tunnel
}

func NewMgr() *Mgr {
	return &Mgr{data: make(map[string]*Tunnel)}
}

func (mgr *Mgr) Add(t *Tunnel) {
	mgr.Lock()
	mgr.data[t.local] = t
	mgr.data[t.remote] = t
	mgr.Unlock()
}
