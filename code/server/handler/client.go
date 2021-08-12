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
	tunnels map[string]struct{}
}

func newClient(parent *Handler, id string, conn *network.Conn) *client {
	return &client{
		parent:  parent,
		id:      id,
		c:       conn,
		tunnels: make(map[string]struct{}),
	}
}

func (c *client) run() {
	for {
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

func (c *client) addTunnel(id string) {
	c.Lock()
	c.tunnels[id] = struct{}{}
	c.Unlock()
}

func (c *client) getTunnels() []string {
	ret := make([]string, 0, len(c.tunnels))
	c.RLock()
	for tn := range c.tunnels {
		ret = append(ret, tn)
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
	delete(c.tunnels, id)
	c.Unlock()
}
