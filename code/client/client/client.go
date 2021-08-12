package client

import (
	"natpass/code/client/global"
	"natpass/code/network"
	"time"

	"github.com/lwch/runtime"
)

type Client struct {
	cfg  *global.Configure
	conn *network.Conn
}

func New(cfg *global.Configure, conn *network.Conn) *Client {
	return &Client{
		cfg:  cfg,
		conn: conn,
	}
}

func (c *Client) Run() {
	err := c.writeHandshake()
	runtime.Assert(err)
	select {}
}

func (c *Client) writeHandshake() error {
	var msg network.Msg
	msg.XType = network.Msg_handshake
	msg.From = c.cfg.ID
	msg.To = "server"
	msg.Payload = &network.Msg_Hsp{
		Hsp: &network.HandshakePayload{
			Enc: c.cfg.Enc[:],
		},
	}
	return c.conn.WriteMessage(&msg, 5*time.Second)
}
