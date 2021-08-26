package handler

import (
	"natpass/code/network"
	"sync"
	"time"
)

type clients struct {
	sync.RWMutex
	parent *Handler
	id     string
	data   map[uint32]*client // idx => client
	idx    int
}

func newClients(parent *Handler, id string) *clients {
	return &clients{
		parent: parent,
		id:     id,
		data:   make(map[uint32]*client),
	}
}

func (cs *clients) new(idx uint32, conn *network.Conn) *client {
	cli := &client{
		parent:  cs,
		idx:     idx,
		conn:    conn,
		updated: time.Now(),
		links:   make(map[string]struct{}),
	}
	cs.Lock()
	cs.data[idx] = cli
	cs.Unlock()
	return cli
}

func (cs *clients) next() *client {
	list := make([]*client, 0, len(cs.data))
	cs.RLock()
	for _, cli := range cs.data {
		list = append(list, cli)
	}
	cs.RUnlock()
	if len(list) > 0 {
		cli := list[cs.idx%len(list)]
		cs.idx++
		return cli
	}
	return nil
}

func (cs *clients) close(idx uint32) {
	cs.Lock()
	delete(cs.data, idx)
	cs.Unlock()
}
