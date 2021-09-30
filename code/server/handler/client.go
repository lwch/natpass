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
	parent  *clients
	idx     uint32
	conn    *network.Conn
	updated time.Time
	links   map[string]struct{} // link id => struct{}
}

func (c *client) run() {
	defer c.parent.parent.closeClient(c)
	for {
		if time.Since(c.updated).Seconds() > 600 {
			links := make([]string, 0, len(c.links))
			c.RLock()
			for id := range c.links {
				links = append(links, id)
			}
			c.RUnlock()
			logging.Info("%s-%d is not keepalived, links: %v", c.parent.id, c.idx, links)
			return
		}
		msg, err := c.conn.ReadMessage(c.parent.parent.cfg.ReadTimeout)
		if err != nil {
			if strings.Contains(err.Error(), "i/o timeout") {
				continue
			}
			logging.Error("read message from %s-%d: %v", c.parent.id, c.idx, err)
			return
		}
		c.updated = time.Now()
		c.parent.parent.onMessage(c, c.conn, msg)
	}
}

func (c *client) writeMessage(msg *network.Msg) error {
	return c.conn.WriteMessage(msg, c.parent.parent.cfg.WriteTimeout)
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

func (c *client) closeLink(id string) {
	var msg network.Msg
	msg.From = "server"
	msg.To = c.parent.id
	msg.ToIdx = c.idx
	msg.XType = network.Msg_disconnect
	msg.LinkId = id
	c.conn.WriteMessage(&msg, c.parent.parent.cfg.WriteTimeout)
	c.Lock()
	delete(c.links, id)
	c.Unlock()
}

func (c *client) closeShell(id string) {
	var msg network.Msg
	msg.From = "server"
	msg.To = c.parent.id
	msg.ToIdx = c.idx
	msg.XType = network.Msg_shell_close
	msg.LinkId = id
	c.conn.WriteMessage(&msg, c.parent.parent.cfg.WriteTimeout)
	c.Lock()
	delete(c.links, id)
	c.Unlock()
}

func (c *client) is(id string, idx uint32) bool {
	return c.parent.id == id && c.idx == idx
}

func (c *client) keepalive() {
	var msg network.Msg
	msg.From = "server"
	msg.To = c.parent.id
	msg.ToIdx = c.idx
	msg.XType = network.Msg_keepalive
	for {
		time.Sleep(10 * time.Second)
		err := c.writeMessage(&msg)
		if err != nil {
			logging.Error("send keepalive: %v", err)
			return
		}
	}
}
