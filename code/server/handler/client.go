package handler

import (
	"strings"
	"sync"
	"time"

	"github.com/lwch/logging"
	"github.com/lwch/natpass/code/network"
)

type client struct {
	sync.RWMutex
	id      string
	parent  *clients
	conn    *network.Conn
	updated time.Time
	links   map[string]struct{} // link id => struct{}
}

func (c *client) close() {
	for _, link := range c.getLinks() {
		c.parent.parent.closeLink(link)
		c.Lock()
		delete(c.links, link)
		c.Unlock()
	}
	c.conn.Close()
	logging.Info("client %s connection closed", c.id)
}

func (c *client) run() {
	defer c.parent.close(c.id)
	for {
		if time.Since(c.updated).Seconds() > 600 {
			links := make([]string, 0, len(c.links))
			c.RLock()
			for id := range c.links {
				links = append(links, id)
			}
			c.RUnlock()
			logging.Info("%s is not keepalived, links: %v", c.id, links)
			return
		}
		msg, size, err := c.conn.ReadMessage(c.parent.parent.cfg.ReadTimeout)
		if err != nil {
			if strings.Contains(err.Error(), "i/o timeout") {
				continue
			}
			logging.Error("read message from %s: %v", c.id, err)
			return
		}
		c.updated = time.Now()
		c.parent.parent.onMessage(c, c.conn, msg, size)
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

func (c *client) sendClose(id string) {
	var msg network.Msg
	msg.From = "server"
	msg.To = c.id
	msg.XType = network.Msg_disconnect
	msg.LinkId = id
	c.conn.WriteMessage(&msg, c.parent.parent.cfg.WriteTimeout)
	c.Lock()
	delete(c.links, id)
	c.Unlock()
}

func (c *client) keepalive() {
	var msg network.Msg
	msg.From = "server"
	msg.To = c.id
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
