package tunnel

import "sync"

type Mgr struct {
	sync.RWMutex
	data map[string]*Tunnel // channel id => tunnel
}

func NewMgr() *Mgr {
	return &Mgr{data: make(map[string]*Tunnel)}
}

func (mgr *Mgr) Add(t *Tunnel) {
	mgr.Lock()
	mgr.data[t.localChannelID] = t
	mgr.data[t.remoteChannelID] = t
	mgr.Unlock()
}

func (mgr *Mgr) Find(id string) *Tunnel {
	mgr.RLock()
	defer mgr.RUnlock()
	return mgr.data[id]
}
