package tunnel

import (
	"fmt"
	"net"

	"github.com/lwch/logging"
)

type Tunnel struct {
	local  string
	remote string
	conn   net.Conn
}

func NewListen(name, local, remote string, conn net.Conn) (*Tunnel, error) {
	logging.Info("new listen: name=%s, addr=%s:%d", name, conn.LocalAddr().String())
	return &Tunnel{
		local:  local,
		remote: remote,
		conn:   conn,
	}, nil
}

func NewConnect(name, local, remote string, t, addr string, port uint32) (*Tunnel, error) {
	logging.Info("new connect: name=%s, remote=%s://%s:%d", name, t, addr, port)
	conn, err := net.Dial(t, fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		return nil, err
	}
	return &Tunnel{
		local:  local,
		remote: remote,
		conn:   conn,
	}, nil
}

func (t *Tunnel) Close() {
	t.conn.Close()
}
