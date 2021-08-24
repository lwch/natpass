package handler

import (
	"errors"
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
	trimID  string
	c       *network.Conn
	links   map[string]struct{} // link id => struct{}
	updated time.Time
}

func newClient(parent *Handler, id, trimID string, conn *network.Conn) *client {
	return &client{
		parent:  parent,
		id:      id,
		trimID:  trimID,
		c:       conn,
		links:   make(map[string]struct{}),
		updated: time.Now(),
	}
}

func (c *client) run() {
	for {
		if time.Since(c.updated).Seconds() > 600 {
			links := make([]string, 0, len(c.links))
			c.RLock()
			for id := range c.links {
				links = append(links, id)
			}
			c.RUnlock()
			logging.Info("%s is not keepalived, links: %v", c.id, links)
			c.parent.closeAll(c)
			return
		}
		msg, err := c.c.ReadMessage(c.parent.cfg.ReadTimeout)
		if err != nil {
			if strings.Contains(err.Error(), "i/o timeout") {
				continue
			}
			logging.Error("read message from %s: %v", c.id, err)
			return
		}
		c.updated = time.Now()
		c.parent.onMessage(c, c.c, msg)
	}
}

func (c *client) writeMessage(msg *network.Msg) error {
	for i := 0; i < 10; i++ {
		err := c.c.WriteMessage(msg, c.parent.cfg.WriteTimeout)
		if err != nil {
			if strings.Contains(err.Error(), "i/o timeout") {
				continue
			}
			return err
		}
	}
	return errors.New("too many retry")
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
	c.c.WriteMessage(&msg, c.parent.cfg.WriteTimeout)
	c.Lock()
	delete(c.links, id)
	c.Unlock()
}
