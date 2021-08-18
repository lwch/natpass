package handler

import (
	"natpass/code/network"
	"strings"
	"sync"
	"time"

	"github.com/lwch/logging"
)

type client struct {
	sync.RWMutex
	parent  *Handler
	id      string
	c       *network.Conn
	links   map[string]struct{} // link id => struct{}
	updated time.Time
}

func newClient(parent *Handler, id string, conn *network.Conn) *client {
	return &client{
		parent:  parent,
		id:      id,
		c:       conn,
		links:   make(map[string]struct{}),
		updated: time.Now(),
	}
}

func (c *client) run() {
	for {
		if time.Since(c.updated).Seconds() > 600 {
			c.parent.closeAll(c)
			return
		}
		msg, err := c.c.ReadMessage(time.Second)
		if err != nil {
			if strings.Contains(err.Error(), "i/o timeout") {
				continue
			}
			logging.Error("read message from %s: %v", c.id, err)
			return
		}
		c.parent.onMessage(msg)
	}
}

func (c *client) writeMessage(msg *network.Msg) error {
	return c.c.WriteMessage(msg, time.Second)
}

func (c *client) addLink(id string) {
	c.Lock()
	c.links[id] = struct{}{}
	c.Unlock()
}

func (c *client) removeLink(id string) {
	c.Lock()
	delete(c.links, id)
	c.Unlock()
}

func (c *client) getLinks() []string {
	ret := make([]string, 0, len(c.links))
	c.RLock()
	for link := range c.links {
		ret = append(ret, link)
	}
	c.RUnlock()
	return ret
}

func (c *client) close(id string) {
	var msg network.Msg
	msg.From = "server"
	msg.To = c.id
	msg.XType = network.Msg_disconnect
	msg.Payload = &network.Msg_XDisconnect{
		XDisconnect: &network.Disconnect{
			Id: id,
		},
	}
	c.c.WriteMessage(&msg, time.Second)
	c.Lock()
	delete(c.links, id)
	c.Unlock()
}
