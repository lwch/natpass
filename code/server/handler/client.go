package handler

import (
	"natpass/code/network"
	"strings"
	"time"

	"github.com/lwch/logging"
)

type client struct {
	id string
	c  *network.Conn
}

func newClient(id string, conn *network.Conn) *client {
	return &client{
		id: id,
		c:  conn,
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
		logging.Info("read message: %s", msg.String())
	}
}
