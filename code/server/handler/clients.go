package handler

import (
	"sync"
	"time"

	"github.com/jkstack/natpass/code/network"
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
	}
	cs.data[id] = cli
	cs.Unlock()
	return cli
}

func (cs *clients) remove(id string) {
	cs.Lock()
	delete(cs.data, id)
	cs.Unlock()
}

func (cs *clients) lookup(id string) *client {
	cs.RLock()
	defer cs.RUnlock()
	return cs.data[id]
}
