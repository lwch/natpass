package client

import (
	"net"

	"github.com/lwch/runtime"
)

type tunnel struct {
	cli     *Client
	id      string
	forward net.Conn
}

func newTunnel(cli *Client, conn net.Conn) *tunnel {
	id, err := runtime.UUID(16, "0123456789abcdef")
	runtime.Assert(err)
	return &tunnel{
		cli:     cli,
		id:      id,
		forward: conn,
	}
}
