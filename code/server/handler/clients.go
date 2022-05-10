package handler

import (
	"sync"
	"time"

	"github.com/lwch/natpass/code/network"
)

type clients struct {
	sync.RWMutex
	parent *Handler
	data   map[string]*client // id => client
}

func newClients(parent *Handler) *clients {
	return &clients{
		parent: parent,
		data:   make(map[string]*client),
	}
}

func (cs *clients) new(id string, conn *network.Conn) *client {
	cli := &client{
		id:      id,
		parent:  cs,
		conn:    conn,
		updated: time.Now(),
		links:   make(map[string]struct{}),
	}
	cs.Lock()
	if c, ok := cs.data[id]; ok {
		c.close()
		delete(cs.data, id)
	}
	cs.data[id] = cli
	cs.Unlock()
	return cli
}

func (cs *clients) lookup(id string) *client {
	cs.RLock()
	defer cs.RUnlock()
	return cs.data[id]
}

func (cs *clients) close(id string) {
	cs.Lock()
	if c, ok := cs.data[id]; ok {
		c.close()
		delete(cs.data, id)
	}
	cs.Unlock()
}
