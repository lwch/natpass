package handler

import (
	"natpass/code/network"
	"strings"
	"time"

	"github.com/lwch/logging"
)

type client struct {
	parent *Handler
	id     string
	c      *network.Conn
}

func newClient(parent *Handler, id string, conn *network.Conn) *client {
	return &client{
		parent: parent,
		id:     id,
		c:      conn,
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
